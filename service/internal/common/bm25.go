package common

import (
	"cmp"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/cespare/xxhash/v2"
	"github.com/go-ego/gse"
)

// BM25 相关参数。
const (
	// bm25Saturation 词频饱和系数 k1，越大词频的影响越持久。
	bm25Saturation = 1.2
	// bm25Weight 长度归一强度 b，0 表示不归一，1 表示完全归一。
	bm25Weight = 0.75
	// bm25Length 全库切片平均字符数 avgdl，作为长度归一的参考值。
	bm25Length = 550.0
)

// stopwords 停用词表，命中后不参与稀疏向量。
var stopwords = map[string]struct{}{
	"的": {}, "了": {}, "在": {}, "是": {}, "我": {}, "有": {}, "和": {}, "就": {},
	"不": {}, "人": {}, "都": {}, "一": {}, "上": {}, "也": {}, "很": {}, "到": {},
	"说": {}, "要": {}, "去": {}, "你": {}, "会": {}, "着": {}, "看": {}, "好": {},
	"这": {}, "那": {}, "他": {}, "她": {}, "它": {}, "们": {}, "与": {}, "及": {},
	"或": {}, "而": {}, "但": {}, "被": {}, "把": {}, "让": {}, "给": {}, "对": {},
	"从": {}, "向": {}, "于": {}, "为": {}, "以": {}, "之": {}, "其": {}, "则": {},
	"乃": {}, "若": {}, "又": {}, "还": {}, "才": {}, "只": {}, "更": {}, "最": {},
	"一个": {}, "自己": {}, "什么": {}, "怎么": {}, "这个": {}, "那个": {}, "我们": {},
	"你们": {}, "他们": {}, "她们": {}, "没有": {}, "就是": {}, "还是": {}, "但是": {},
	"然后": {}, "以及": {}, "因为": {}, "所以": {}, "如果": {}, "虽然": {}, "不过": {},
	"可以": {}, "已经": {}, "于是": {}, "并且": {}, "而且": {}, "或者": {}, "只是": {},
}

// 分词器及其状态。
var (
	// segmenter 全局分词器，gse 不支持并发使用，由 segmenterLock 保护。
	segmenter gse.Segmenter
	// segmenterLock 分词器并发锁。
	segmenterLock sync.Mutex
)

// term 词下标及其权重。
type term struct {
	// index 词下标。
	index uint32
	// weight 词权重。
	weight float32
}

// Initialize 加载分词词典。
//
// return: 错误信息。
func Initialize() error {
	var err error
	// 加载内置词典。
	if segmenter, err = gse.NewEmbed("zh"); err != nil {
		slog.Error("[BM25]加载分词词典失败", "error", err)
		return err
	}
	// 返回成功结果。
	return nil
}

// GetMatching 计算文本的稀疏向量（不含 IDF，IDF 由向量库在查询时补上）。
//
// param: content 切片文本。
// return: 词下标，词权重。
func GetMatching(content string) ([]uint32, []float32) {
	words := tokenize(content)
	if len(words) == 0 {
		return nil, nil
	}

	// 长度按字符计，与分词结果的词数口径无关，仅作为长度归一的近似值。
	length := float64(utf8.RuneCountInString(content))
	return weight(words, length)
}

// tokenize 分词并清洗文本。
//
// param: text 待分词文本。
// return: 词列表，保留顺序与重复。
func tokenize(text string) []string {
	segmenterLock.Lock()
	words := segmenter.Cut(text, true)
	segmenterLock.Unlock()

	result := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.ToLower(word)
		if word == "" || isSymbol(word) {
			continue
		}

		if _, ok := stopwords[word]; ok {
			continue
		}

		result = append(result, word)
	}
	return result
}

// isSymbol 判断词是否仅由符号组成。
//
// param: word 词。
// return: 是否仅由符号组成。
func isSymbol(word string) bool {
	for _, letter := range word {
		if unicode.IsLetter(letter) || unicode.IsNumber(letter) {
			return false
		}
	}
	return true
}

// weight 统计词频并计算权重。
//
// param: words 词列表。
// param: length 文本字符数。
// return: 词下标，词权重。
func weight(words []string, length float64) ([]uint32, []float32) {
	counts := make(map[string]int, len(words)/2)
	for _, word := range words {
		counts[word]++
	}

	norm := 1 - bm25Weight + bm25Weight*length/bm25Length
	terms := make([]term, 0, len(counts))
	for word, frequency := range counts {
		score := float64(frequency) * (bm25Saturation + 1) /
			(float64(frequency) + bm25Saturation*norm)
		terms = append(terms, term{index: uint32(xxhash.Sum64String(word)), weight: float32(score)})
	}

	// 按下标排序，保证同样的输入得到完全相同的输出。
	slices.SortFunc(terms, func(left term, right term) int {
		return cmp.Compare(left.index, right.index)
	})

	indices := make([]uint32, len(terms))
	values := make([]float32, len(terms))
	for i, item := range terms {
		indices[i] = item.index
		values[i] = item.weight
	}
	return indices, values
}
