package pipeline

import (
	"context"
	"fmt"
	"regexp"
	"service/internal/storage"
	"strings"
	"unicode/utf8"
	"uuid"
)

// segmentMinimum 分片长度下限（字符）。
const segmentMinimum = 300

// segmentMaximum 分片长度上限（字符）。
const segmentMaximum = 600

// overlapMinimum 重叠长度下限（字符）。
const overlapMinimum = 50

// overlapMaximum 重叠长度上限（字符）。
const overlapMaximum = 100

// fallbackTitle 未识别到章节时使用的标题。
const fallbackTitle = "全文"

// chapterPattern 中文章节标题匹配规则（用于没有 Markdown 标题的文本）。
var chapterPattern = regexp.MustCompile(`^第[0-9０-９〇零一二三四五六七八九十百千万两]+[章节回卷篇]`)

// englishPattern 英文章节标题匹配规则。
var englishPattern = regexp.MustCompile(`(?i)^(chapter|part)\s+[0-9]+`)

// chapter 章节信息。
type chapter struct {
	// title 章节标题。
	title string
	// start 起始位置（归一化文档的字符下标）。
	start int32
	// final 结束位置（归一化文档的字符下标）。
	final int32
	// paragraphs 章节正文段落。
	paragraphs []paragraph
}

// paragraph 段落信息。
type paragraph struct {
	// text 段落正文。
	text string
	// start 起始位置（归一化文档的字符下标）。
	start int32
	// final 结束位置（归一化文档的字符下标）。
	final int32
}

// fragment 分片信息。
type fragment struct {
	// prefix 与上一分片重叠的内容，首片为空。
	prefix *string
	// middle 分片独有内容。
	middle string
	// suffix 与下一分片重叠的内容，末片为空。
	suffix *string
	// point1 前缀起始位置（归一化文档的字符下标）。
	point1 int32
	// point2 正文起始位置（归一化文档的字符下标）。
	point2 int32
	// point3 后缀起始位置（归一化文档的字符下标）。
	point3 int32
	// point4 分片结束位置（归一化文档的字符下标）。
	point4 int32
}

// organize 分章切片文件。
//
// param: ctx 上下文。
// param: origin 原始文件编码。
// param: novel 小说编码。
// return: 分片总数，字符数，章节数，错误信息。
func organize(ctx context.Context, origin string, novel string) (int32, int32, int32, error) {
	// 读取归一化文档。
	data, err := storage.OSS.Select(ctx, "Normal", origin+".md")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("读取归一化文档失败，详细错误信息：%w", err)
	}
	// 识别章节并统计字符数。
	text := string(data)
	chapters := divide(text)
	characters := int32(utf8.RuneCountInString(text))
	// 清理已有数据，保证重复执行结果一致。
	if err := storage.RDB.QueryNaught(ctx, `delete from segment where novel=$1`, novel); err != nil {
		return 0, 0, 0, fmt.Errorf("清理分片失败，详细错误信息：%w", err)
	}
	if err := storage.RDB.QueryNaught(ctx, `delete from chapter where novel=$1`, novel); err != nil {
		return 0, 0, 0, fmt.Errorf("清理章节失败，详细错误信息：%w", err)
	}
	if err := storage.VDB.DeletePoints(ctx, "novel", novel); err != nil {
		return 0, 0, 0, fmt.Errorf("清理向量失败，详细错误信息：%w", err)
	}
	// 逐章保存并切片。
	var total int32
	for index, item := range chapters {
		count, err := store(ctx, novel, int32(index+1), item, total)
		if err != nil {
			return 0, 0, 0, err
		}
		total += count
	}
	// 返回分片总数、字符数、章节数。
	return total, characters, int32(len(chapters)), nil
}

// divide 识别章节并拆分段落。
//
// param: text 归一化文档。
// return: 章节列表。
func divide(text string) []chapter {
	lines := strings.Split(text, "\n")
	// 计算每行的字符起点。
	starts := make([]int32, len(lines))
	var offset int32
	for index, line := range lines {
		starts[index] = offset
		offset += int32(utf8.RuneCountInString(line)) + 1
	}
	total := int32(utf8.RuneCountInString(text))
	// 标记章节起始行与标题。
	marks := make([]int, 0)
	titles := make([]string, 0)
	for index, line := range lines {
		name, exists := heading(line)
		if exists == false {
			name, exists = match(line)
		}
		if exists {
			marks = append(marks, index)
			titles = append(titles, name)
		}
	}
	// 未识别到章节时整本作为一章。
	if len(marks) == 0 {
		return []chapter{{title: fallbackTitle, start: 0, final: total, paragraphs: collect(lines, starts, 0, len(lines))}}
	}
	// 按章节标记拆分。
	result := make([]chapter, 0, len(marks))
	for index, mark := range marks {
		// 计算章节范围与正文范围。
		begin := starts[mark]
		from := mark + 1
		final := total
		to := len(lines)
		if index+1 < len(marks) {
			final = starts[marks[index+1]]
			to = marks[index+1]
		}
		// 首个章节前的导言并入首章。
		paragraphs := collect(lines, starts, from, to)
		if index == 0 && mark > 0 {
			begin = 0
			paragraphs = append(collect(lines, starts, 0, mark), paragraphs...)
		}
		result = append(result, chapter{title: titles[index], start: begin, final: final, paragraphs: paragraphs})
	}
	return result
}

// heading 判断是否为 Markdown 标题行。
//
// param: line 文本行。
// return: 章节标题，是否为标题行。
func heading(line string) (string, bool) {
	text := strings.TrimSpace(line)
	if strings.HasPrefix(text, "#") == false {
		return "", false
	}
	// 跳过井号。
	index := 0
	for index < len(text) && text[index] == '#' {
		index++
	}
	if index == len(text) {
		return "", false
	}
	// 井号后需要有分隔，避免把 #话题 当成标题。
	if text[index] != ' ' && text[index] != '\t' {
		return "", false
	}
	return title(strings.TrimSpace(text[index:])), true
}

// match 判断是否为章节标题行。
//
// param: line 文本行。
// return: 章节标题，是否为标题行。
func match(line string) (string, bool) {
	text := strings.TrimSpace(line)
	if chapterPattern.MatchString(text) || englishPattern.MatchString(text) {
		return title(text), true
	}
	return "", false
}

// title 截取合法的章节标题。
//
// param: text 章节标题。
// return: 长度不超过 200 字符的章节标题。
func title(text string) string {
	runes := []rune(text)
	if len(runes) > 200 {
		return string(runes[:200])
	}
	return text
}

// collect 收集指定行区间内的段落。
//
// param: lines 文档行列表。
// param: starts 每行的字符起点。
// param: from 起始行下标。
// param: to 结束行下标（不含）。
// return: 段落列表。
func collect(lines []string, starts []int32, from int, to int) []paragraph {
	result := make([]paragraph, 0)
	for index := from; index < to && index < len(lines); index++ {
		text := strings.TrimSpace(lines[index])
		if text == "" {
			continue
		}
		// 计算去除首部空白后的字符起点。
		lead := int32(utf8.RuneCountInString(lines[index][:strings.Index(lines[index], text)]))
		start := starts[index] + lead
		result = append(result, paragraph{text: text, start: start, final: start + int32(utf8.RuneCountInString(text))})
	}
	return result
}

// section 按段落切片，相邻分片保留末尾若干段落作为重叠。
//
// param: paragraphs 章节段落列表。
// return: 分片列表。
func section(paragraphs []paragraph) []fragment {
	groups := make([][2]int, 0)
	start := 0
	// lower 本片净片起点，下一片的重叠不能越过它。
	lower := 0
	for start < len(paragraphs) {
		// 向后聚合段落，直至达到长度上限。
		end := start
		size := 0
		for end < len(paragraphs) {
			length := measure(paragraphs[end])
			if size >= segmentMinimum && size+length > segmentMaximum {
				break
			}
			size += length
			end++
			if size >= segmentMaximum {
				break
			}
		}
		groups = append(groups, [2]int{start, end})
		if end >= len(paragraphs) {
			break
		}
		// 回退若干段落作为下一片的开头，不越过本片净片起点。
		start = rewind(paragraphs, lower, end)
		lower = end
	}
	// 章节末尾过短的片并入上一片。
	if len(groups) > 1 {
		last := groups[len(groups)-1]
		if rangeMeasure(paragraphs, last[0], last[1]) < segmentMinimum {
			previous := groups[len(groups)-2]
			if rangeMeasure(paragraphs, previous[0], last[1]) <= segmentMaximum {
				groups = groups[:len(groups)-1]
				groups[len(groups)-1][1] = last[1]
			}
		}
	}
	// 生成分片。
	result := make([]fragment, 0, len(groups))
	for index, group := range groups {
		begin := group[0]
		end := group[1]
		// 净片起点：首片即分片起点，其余为上一分片的结尾。
		middleStart := begin
		if index > 0 {
			middleStart = groups[index-1][1]
		}
		// 净片终点：末片即分片结尾，其余为下一分片的起点。
		middleFinal := end
		if index+1 < len(groups) {
			middleFinal = groups[index+1][0]
		}
		if middleStart > middleFinal {
			middleStart = middleFinal
		}
		// 组装分片内容。
		item := fragment{middle: combine(paragraphs, middleStart, middleFinal)}
		if middleStart > begin {
			text := combine(paragraphs, begin, middleStart)
			item.prefix = &text
		}
		if middleFinal < end {
			text := combine(paragraphs, middleFinal, end)
			item.suffix = &text
		}
		// 记录三个分段的文档位置。
		item.point2 = paragraphs[middleStart].start
		item.point1 = item.point2
		if item.prefix != nil {
			item.point1 = paragraphs[begin].start
		}
		item.point4 = paragraphs[end-1].final
		item.point3 = item.point4
		if item.suffix != nil {
			item.point3 = paragraphs[middleFinal].start
		}
		result = append(result, item)
	}
	return result
}

// combine 拼接指定段落区间的正文。
//
// param: paragraphs 章节段落列表。
// param: from 起始段落下标。
// param: to 结束段落下标（不含）。
// return: 以换行符分隔的正文。
func combine(paragraphs []paragraph, from int, to int) string {
	texts := make([]string, 0, to-from)
	for index := from; index < to; index++ {
		texts = append(texts, paragraphs[index].text)
	}
	return strings.Join(texts, "\n")
}

// rewind 计算下一片的起始段落，保留末尾若干段落作为重叠。
//
// param: paragraphs 章节段落列表。
// param: start 本片净片起点，重叠不能越过该位置。
// param: end 本片结束段落下标。
// return: 下一片的起始段落下标。
func rewind(paragraphs []paragraph, start int, end int) int {
	next := end
	total := 0
	for next > start {
		length := measure(paragraphs[next-1])
		// 超出重叠上限时停止，首段允许超出。
		if total > 0 && total+length > overlapMaximum {
			break
		}
		total += length
		next--
		if total >= overlapMinimum {
			break
		}
	}
	// 重叠覆盖整片时放弃重叠，保证处理向前推进。
	if next <= start {
		return end
	}
	return next
}

// measure 统计段落的字符数。
//
// param: item 段落信息。
// return: 段落的字符数。
func measure(item paragraph) int {
	return utf8.RuneCountInString(item.text)
}

// rangeMeasure 统计指定段落区间的字符数。
//
// param: paragraphs 章节段落列表。
// param: from 起始段落下标。
// param: to 结束段落下标（不含）。
// return: 区间内的字符数。
func rangeMeasure(paragraphs []paragraph, from int, to int) int {
	total := 0
	for index := from; index < to; index++ {
		total += measure(paragraphs[index])
	}
	return total
}

// store 保存单个章节及其分片。
//
// param: ctx 上下文。
// param: novel 小说编码。
// param: serial 章节序号。
// param: item 章节信息。
// param: offset 已有分片数量，用于计算分片序号。
// return: 本章节的分片数量，错误信息。
func store(ctx context.Context, novel string, serial int32, item chapter, offset int32) (int32, error) {
	// 计算分片。
	fragments := section(item.paragraphs)
	// 启动数据事务。
	var batch storage.Batch
	// 插入章节记录。
	identifier := strings.ToUpper(uuid.NewV4().String())
	batch.Append("insert into chapter(id,novel,serial,title,start,final,character_count) values($1,$2,$3,$4,$5,$6,$7);", identifier, novel, serial, item.title, item.start, item.final, item.final-item.start)
	// 插入分片记录。
	for index, fragment := range fragments {
		batch.Append("insert into segment(id,novel,chapter,serial,point1,prefix,point2,middle,point3,suffix,point4) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11);", strings.ToUpper(uuid.NewV4().String()), novel, identifier, offset+int32(index)+1, fragment.point1, fragment.prefix, fragment.point2, fragment.middle, fragment.point3, fragment.suffix, fragment.point4)
	}
	// 提交数据事务。
	err := storage.RDB.BatchNaught(ctx, &batch)
	if err != nil {
		return 0, err
	}
	return int32(len(fragments)), nil
}

// prefix 生成分片正文的章节前缀。
//
// param: serial 章节序号。
// param: name 章节标题。
// return: 以换行符结尾的章节前缀。
func prefix(serial int32, name string) string {
	// 标题本身已含章节标记时不再重复编号。
	if name == fallbackTitle || chapterPattern.MatchString(name) || englishPattern.MatchString(name) {
		return name + "\n"
	}
	return fmt.Sprintf("第%d章 %s\n", serial, name)
}
