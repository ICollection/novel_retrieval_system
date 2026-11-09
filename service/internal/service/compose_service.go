package service

import (
	"errors"
	"fmt"
	"service/internal/common"
	"service/internal/config"
	"service/internal/server"
	"service/internal/storage"
	"strings"
	"time"
	"uuid"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// composeService 问答编排服务。
type composeService struct {
}

// ComposeAnswer 编排回答问题。
//
// param: ctx 上下文。
// param: session 会话编码。
// param: content 用户问题。
// param: novels 参考小说。
// param: callback 回调函数。
// return: 会话编码，错误信息。
func (cs *composeService) ComposeAnswer(ctx *gin.Context, session *string, content *string, novels []string, callback func(kind string, content any) error) (*string, error) {
	var consume consume
	// 保存用户问题。
	session, err := cs.appendHumanMessage(ctx, session, content, novels)
	if err != nil {
		return nil, err
	}
	// 召回相关切片。
	question, err := cs.makeupQueryQuestion(ctx, *session, &consume)
	if err != nil {
		return nil, err
	}
	var recalls []recall
	if question == nil {
		recalls, err = cs.recallThroughRecent(ctx, *session)
	} else {
		recalls, err = cs.recallThroughVector(ctx, question, novels)
	}
	if err != nil {
		return nil, err
	}
	// 组织回答问题。
	prompts := []string{}
	prompt1, references, err := cs.makeupConsultPrompt(ctx, recalls)
	if err != nil {
		return nil, err
	}
	prompts = append(prompts, prompt1...)
	prompt2, err := cs.makeupSessionPrompt(ctx, *session)
	if err != nil {
		return nil, err
	}
	prompts = append(prompts, prompt2...)
	answer, llmPrompt, llmAnswer, err := storage.LLM.AnswerStream(ctx, config.Normal, strings.Join(prompts, "\n"), callback)
	consume.llmPrompt = llmPrompt
	consume.llmAnswer = llmAnswer
	for reference := range references {
		callback("refer", reference)
	}
	if err != nil {
		cs.appendModelMessage(ctx, *session, answer, novels, lo.Map(recalls, func(item recall, _ int) string { return item.ID }), consume, false)
		return nil, err
	}
	// 保存问题答案。
	err = cs.appendModelMessage(ctx, *session, answer, novels, lo.Map(recalls, func(item recall, _ int) string { return item.ID }), consume, true)
	if err != nil {
		return nil, err
	}
	// 返回会话编码。
	return session, nil
}

// appendHumanMessage 保存用户问题。
//
// param: ctx 上下文。
// param: session 会话编码。
// param: content 用户问题。
// param: novels 参考小说。
// return: 会话编码，错误消息。
func (cs *composeService) appendHumanMessage(ctx *gin.Context, session *string, content *string, novels []string) (*string, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 验证会话状态。
	if session != nil {
		user, err := storage.RDB.QuerySimple[string](ctx, "select create_user from session where id=$1;", *session)
		if err != nil {
			return nil, err
		}
		if user == nil || *user != claims.Subject {
			return nil, errors.New("会话不属于此用户。")
		}
	}
	// 是否已有会话。
	if session == nil {
		// 组织会话消息。
		moment := time.Now().UTC()
		session := strings.ToUpper(uuid.NewV4().String())
		title := common.Truncate(*content, 50)
		serial := 1
		// 保存会话消息。
		var batch storage.Batch
		batch.Append("insert into session(id,title,summary,contain,current,create_user,create_time,update_time) values($1,$2,$3,$4,$5,$6,$7,$8);", session, title, nil, nil, serial, claims.Subject, moment, moment)
		batch.Append("insert into message(id,session,serial,human,content,novels,segments,complete,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", strings.ToUpper(uuid.NewV4().String()), session, serial, true, content, strings.Join(novels, ","), nil, true, moment)
		err := storage.RDB.BatchNaught(ctx, &batch)
		if err != nil {
			return nil, err
		}
		// 返回消息编码。
		return &session, nil
	} else {
		moment := time.Now().UTC()
		// 获取当前序号。
		current, err := storage.RDB.QuerySimple[int](ctx, "select current from session where id=$1", session)
		if err != nil {
			return session, err
		}
		serial := *current + 1
		// 保存会话消息。
		var batch storage.Batch
		batch.Append("update session set current=$1,update_time=$2 where id=$3;", serial, moment, session)
		batch.Append("insert into message(id,session,serial,human,content,novels,segments,complete,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", strings.ToUpper(uuid.NewV4().String()), session, serial, true, content, strings.Join(novels, ","), nil, true, moment)
		err = storage.RDB.BatchNaught(ctx, &batch)
		if err != nil {
			return nil, err
		}
		// 压缩历史摘要。

		// 返回消息编码。
		return session, nil
	}
}

// makeupQueryQuestion 构建查询问题。
//
// param: ctx 上下文。
// param: session 会话编码。
// param: consume 消费词元。
// return: 问题向量，错误信息。
func (cs *composeService) makeupQueryQuestion(ctx *gin.Context, session string, consume *consume) ([]float32, error) {
	// 获取查询文本。
	prompt, err := cs.makeupSessionPrompt(ctx, session)
	if err != nil {
		return nil, err
	}
	// 小模型整理问题。
	question, slmPrompt, slmAnswer, err := storage.SLM.AnswerDirect(ctx, config.Recall, strings.Join(prompt, "\n"))
	consume.slmPrompt = slmPrompt
	consume.slmAnswer = slmAnswer
	if err != nil {
		return nil, err
	}
	// 维持原有问题。
	if *question == "CONTINUE" {
		return nil, nil
	}
	// 返回问题向量。
	result, vemPrompt, err := storage.VEM.ComputeVector(ctx, *question)
	consume.vemPrompt = vemPrompt
	if err != nil {
		return nil, err
	}
	return result, nil
}

// makeupSessionPrompt 组织会话提示词。
//
// param: ctx 上下文。
// param: string 会话编码。
// return: 提示词行，错误消息。
func (cs *composeService) makeupSessionPrompt(ctx *gin.Context, id string) ([]string, error) {
	result := []string{}
	// 获取会话记录。
	session, err := storage.RDB.QuerySingle[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where id=$1", id)
	if err != nil {
		return nil, err
	}
	// 加入历史摘要。
	if session.Contain != nil {
		result = append(result, "======== 以下为历史摘要 ========")
		result = append(result, "历史摘要："+*session.Summary)
	}
	// 加入历史问题。
	var messages []message
	if session.Contain != nil {
		messages, err = storage.RDB.QueryMatrix[message](ctx, "select id,session,serial,human,content,novels,segments,complete,create_time from message where session=$1 and serial<$2 and serial>$3 order by serial;", session.ID, session.Current, session.Contain)
	} else {
		messages, err = storage.RDB.QueryMatrix[message](ctx, "select id,session,serial,human,content,novels,segments,complete,create_time from message where session=$1 and serial<$2 order by serial;", session.ID, session.Current)
	}
	if err != nil {
		return nil, err
	}
	if len(messages) > 0 {
		result = append(result, "======== 以下为历史对话 ========")
	}
	for _, message := range messages {
		if *message.Human {
			result = append(result, "用户提问："+*message.Content)
		} else {
			result = append(result, "助手回答："+*message.Content)
		}
	}
	// 加入当前问题。
	message, err := storage.RDB.QuerySingle[message](ctx, "select id,session,serial,human,content,novels,segments,complete,create_time from message where session=$1 and serial=$2 order by serial;", session.ID, session.Current)
	if err != nil {
		return nil, err
	}
	result = append(result, "======== 以下为当前问题 ========")
	result = append(result, "用户提问："+*message.Content)
	// 返回合成消息。
	return result, nil
}

// recallThroughRecent 通过记录召回切片。
//
// param: ctx 上下文。
// param: session 会话编码。
// return: 上次切片，错误信息。
func (cs *composeService) recallThroughRecent(ctx *gin.Context, session string) ([]recall, error) {
	// 查询上次切片。
	ids, err := storage.RDB.QuerySimple[string](ctx, "select segments from message where session=$1 and segments is not null order by serial desc limit 1;", session)
	if err != nil {
		return nil, nil
	}
	if ids == nil {
		return []recall{}, nil
	}
	// 获取切片信息。
	recalls, err := storage.RDB.QueryMatrix[recall](ctx, fmt.Sprintf("select id,novel,chapter,serial,0 as score from segment where id in ('%s');", strings.ReplaceAll(*ids, ",", "','")))
	if err != nil {
		return nil, err
	}
	return recalls, nil
}

// recallThroughVector 通过向量召回切片。
//
// param: ctx 上下文。
// param: question 问题向量。
// param: novels 小说范围。
// return: 召回切片，错误信息。
func (cs *composeService) recallThroughVector(ctx *gin.Context, question []float32, novels []string) ([]recall, error) {
	points, err := storage.VDB.FilterPoints(ctx, question, novels, 20, 0.5)
	if err != nil {
		return nil, err
	}
	recalls := make([]recall, 0, len(points))
	for _, point := range points {
		recalls = append(recalls, recall{
			ID:      strings.ToUpper(point.Id.GetUuid()),
			Score:   point.Score,
			Novel:   point.Payload["novel"].GetStringValue(),
			Chapter: point.Payload["chapter"].GetStringValue(),
			Serial:  int32(point.Payload["serial"].GetIntegerValue()),
		})
	}
	return recalls, nil
}

// prompts 获取提问用参考提示词。
//
// param: ctx 上下文。
// param: recalls 召回切片。
// return: 提示词行，参考资料，错误信息。
func (cs *composeService) makeupConsultPrompt(ctx *gin.Context, recalls []recall) ([]string, []reference, error) {
	result := []string{}
	references := []reference{}
	// 合并相关区片。
	sections := sections{sections: []section{}}
	for _, recall := range recalls {
		sections.append(recall.Novel, recall.Chapter, recall.Serial)
	}
	// 查询拼写提示。
	for i, section := range sections.sections {
		var builder strings.Builder
		segments, err := storage.RDB.QueryMatrix[segment](ctx, "select segment.id as segment_id,novel.id as novel_id,novel.title as novel_title,chapter.id as chapter_id,chapter.title as chapter_title,segment.prefix,segment.middle,segment.suffix from segment inner join novel on novel.id=segment.novel inner join chapter on chapter.id=segment.chapter where segment.novel=$1 and segment.chapter=$2 and segment.serial>=$3 and segment.serial<=$4 order by segment.serial;", section.novel, section.chapter, section.start, section.final)
		if err != nil || len(segments) == 0 {
			return nil, nil, err
		}
		for j, segment := range segments {
			if j == 0 && segment.Prefix != nil {
				builder.WriteString(*segment.Prefix)
				builder.WriteString("\n")
			}
			builder.WriteString(*segment.Middle)
			builder.WriteString("\n")
			if segment.Suffix != nil {
				builder.WriteString(*segment.Suffix)
				builder.WriteString("\n")
			}
		}
		result = append(result, fmt.Sprintf("====== 参考片段【%d】《%s》%s ======", i+1, segments[0].NovelTitle, segments[0].ChapterTitle))
		result = append(result, builder.String())
		references = append(references, reference{Serial: i + 1, NovelId: segments[0].NovelID, NovelTitle: segments[0].NovelTitle, ChapterId: segments[0].ChapterID, ChapterTitle: segments[0].ChapterTitle})
	}
	return result, references, nil
}

// appendModelMessage 保存模型答案。
//
// param: ctx 上下文。
// param: session 会话编码。
// param: content 用户问题。
// param: novels 参考小说。
// param: segments 切片编码。
// param: consume 消费词元。
// param: complete 是否完成。
// return: 错误消息。
func (cs *composeService) appendModelMessage(ctx *gin.Context, session string, content *string, novels []string, segments []string, consume consume, complete bool) error {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 组织会话消息。
	moment := time.Now().UTC()
	message := strings.ToUpper(uuid.NewV4().String())
	current, err := storage.RDB.QuerySimple[int](ctx, "select current from session where id=$1", session)
	if err != nil {
		return err
	}
	serial := *current + 1
	// 保存会话消息。
	var batch storage.Batch
	batch.Append("update session set current=$1,update_time=$2 where id=$3;", serial, moment, session)
	batch.Append("insert into message(id,session,serial,human,content,novels,segments,complete,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", message, session, serial, false, content, strings.Join(novels, ","), common.Trim(strings.Join(segments, ",")), complete, moment)
	batch.Append("insert into consume(id,account,session,message,model,prompt,answer) values($1,$2,$3,$4,$5,$6,$7)", strings.ToUpper(uuid.NewV7().String()), claims.Subject, session, message, "SLM", consume.slmPrompt, consume.slmAnswer)
	if consume.vemPrompt > 0 {
		batch.Append("insert into consume(id,account,session,message,model,prompt,answer) values($1,$2,$3,$4,$5,$6,$7)", strings.ToUpper(uuid.NewV7().String()), claims.Subject, session, message, "VEM", consume.vemPrompt, 0)
	}
	batch.Append("insert into consume(id,account,session,message,model,prompt,answer) values($1,$2,$3,$4,$5,$6,$7)", strings.ToUpper(uuid.NewV7().String()), claims.Subject, session, message, "LLM", consume.llmPrompt, consume.llmAnswer)
	err = storage.RDB.BatchNaught(ctx, &batch)
	if err != nil {
		return err
	}
	return nil
}

// 消费词元。
type consume struct {
	// llmPrompt 大语言模型输入数。
	llmPrompt int32
	// llmAnswer 大语言模型输出数。
	llmAnswer int32
	// slmPrompt 小语言模型输入数。
	slmPrompt int32
	// slmAnswer 小语言模型输出数。
	slmAnswer int32
	// vemPrompt 向量模型输入数。
	vemPrompt int32
}

// recall 召回切片。
type recall struct {
	// ID 切片编码。
	ID string `db:"id"`
	// 所属小说。
	Novel string `db:"novel"`
	// 所属章节。
	Chapter string `db:"chapter"`
	// 切片序号。
	Serial int32 `db:"serial"`
	// 召回评分。
	Score float32 `db:"score"`
}

// reference 参考资料。
type reference struct {
	// Serial 资料序号。
	Serial int `json:"serial"`
	// NovelId 小说编码。
	NovelId *string `json:"novel_id"`
	// NovelTitle 小说名称。
	NovelTitle *string `json:"novel_title"`
	// ChapterId 章节编码。
	ChapterId *string `json:"chapter_id"`
	// ChapterTitle 章节名称。
	ChapterTitle *string `json:"chapter_title"`
}

// 切片区间。
type section struct {
	// novel 小说编码。
	novel string
	// chapter 章节编码。
	chapter string
	// start 开始序号。
	start int32
	// final 结束序号。
	final int32
}

// 切片区间集合。
type sections struct {
	// 切片区间。
	sections []section
}

// append 添加选出切片。
//
// param: novel 小说编码。
// param: chapter 章节编码。
// param: serial 切片序号。
func (sections *sections) append(novel string, chapter string, serial int32) {
	index := -1
	// 附加选出切片。
	for i := range sections.sections {
		section := &sections.sections[i]
		if section.novel == novel && section.chapter == chapter {
			if serial > section.start && serial < section.final {
				index = i
				break
			} else if serial >= section.start-2 && serial <= section.start {
				section.start = serial - 1
				index = i
				break
			} else if serial >= section.final && serial <= section.final+2 {
				section.final = serial + 1
				index = i
				break
			}
		}
	}
	// 没有附加成功。
	if index == -1 {
		sections.sections = append(sections.sections, section{novel: novel, chapter: chapter, start: serial - 1, final: serial + 1})
		return
	}
	// 合并分片区间。
	for i := range sections.sections {
		section := &sections.sections[i]
		if index != i && section.novel == novel && section.chapter == chapter {
			current := &sections.sections[index]
			if current.final >= section.start-1 && current.start < section.final {
				current.final = section.final
				sections.sections = append(sections.sections[:i], sections.sections[i+1:]...)
				return
			} else if current.start <= section.final+1 && current.final > section.start {
				current.start = section.start
				sections.sections = append(sections.sections[:i], sections.sections[i+1:]...)
				return
			}
		}
	}
}

// segment 切片数据。
type segment struct {
	// SegmentID 切片编码。
	SegmentID *string `db:"segment_id"`
	// NovelID 小说编码。
	NovelID *string `db:"novel_id"`
	// NovelTitle 小说名称。
	NovelTitle *string `db:"novel_title"`
	// ChapterID 章节编码
	ChapterID *string `db:"chapter_id"`
	// ChapterTitle 章节名称。
	ChapterTitle *string `db:"chapter_title"`
	// Prefix 前缀文本。
	Prefix *string `db:"prefix"`
	// Middle 切片文本。
	Middle *string `db:"middle"`
	// Suffix 后缀文本。
	Suffix *string `db:"suffix"`
}
