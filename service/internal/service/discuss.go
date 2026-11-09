package service

import (
	"errors"
	"service/internal/server"
	"service/internal/storage"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// discussService 用户问答服务。
type discussService struct {
}

// FilterSession 查询用户会话。
//
// param: ctx 上下文。
// return: 用户会话，错误信息。
func (ds *discussService) FilterSession(ctx *gin.Context) ([]session, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 查询用户会话。
	result, err := storage.RDB.QueryMatrix[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where create_user=$1 order by update_time desc;", claims.Subject)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ModifySession 修改用户会话。
//
// param: ctx 上下文。
// param: id 会话编码。
// param: title 对话标题。
// return: 用户会话，错误信息。
func (ds *discussService) ModifySession(ctx *gin.Context, id string, title string) (*string, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 验证会话存在。
	session, err := storage.RDB.QuerySingle[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where id=$1", id)
	if err != nil {
		return nil, err
	}
	if *session.CreateUser != claims.Subject {
		return nil, errors.New("会话不是用户创建。")
	}
	// 修改用户会话。
	err = storage.RDB.QueryNaught(ctx, "update session set title=$1 where id=$2 ", title, id)
	if err != nil {
		return nil, err
	}
	// 返回会话编码。
	return session.ID, nil
}

// DeleteSession 删除用户会话。
//
// param: ctx 上下文。
// param: id 会话编码。
// return: 会话编码，错误信息。
func (ds *discussService) DeleteSession(ctx *gin.Context, id string) (*string, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 验证会话存在。
	session, err := storage.RDB.QuerySingle[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where id=$1", id)
	if err != nil {
		return nil, err
	}
	if *session.CreateUser != claims.Subject {
		return nil, errors.New("会话不是用户创建。")
	}
	// 删除用户会话。
	var batch storage.Batch
	batch.Append("delete from message where session=$1", session.ID)
	batch.Append("delete from session where id=$1", session.ID)
	err = storage.RDB.BatchNaught(ctx, &batch)
	if err != nil {
		return nil, err
	}
	// 返回会话编码。
	return session.ID, nil
}

// FilterMessage 查找会话消息。
//
// param: ctx 上下文。
// param: id 会话编码。
// param: recent 近期数量。
// return: 会话消息，错误信息。
func (ds *discussService) FilterMessage(ctx *gin.Context, id *string, recent int32) ([]message, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 获取会话记录。
	session, err := storage.RDB.QuerySingle[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where id=$1", id)
	if err != nil {
		return nil, err
	}
	if session == nil || *session.CreateUser != claims.Subject {
		return nil, errors.New("会话不存在或不属于用户。")
	}
	// 获取最近会话。
	messages, err := storage.RDB.QueryMatrix[message](ctx, "select id,session,serial,human,content,novels,segments,complete,create_time from message where session=$1 and serial<=$2 and serial>$3 order by serial;", session.ID, session.Current, *session.Current-recent)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// FilterNovels 查询会话小说。
//
// param: ctx 上下文。
// param: id 会话编码。
// return: 小说编码，错误信息。
func (ds *discussService) FilterNovels(ctx *gin.Context, id *string) ([]string, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 获取会话记录。
	session, err := storage.RDB.QuerySingle[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where id=$1", id)
	if err != nil {
		return nil, err
	}
	if session == nil || *session.CreateUser != claims.Subject {
		return nil, errors.New("会话不存在或不属于用户。")
	}
	// 查询会话小说。
	novels, err := storage.RDB.QuerySimple[string](ctx, "select novels from message where novels is not null and session=$1 order by serial desc limit 1;", session.ID)
	if err != nil {
		return nil, err
	}
	// 返回会话小说。
	return strings.Split(*novels, ","), nil
}

// ConsumeMission 统计会话词元。
//
// param: ctx 上下文。
// param: id 会话编码。
// return: 消费词元，错误信息。
func (ds *discussService) ConsumeMission(ctx *gin.Context, id *string) (*consume, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 获取会话记录。
	session, err := storage.RDB.QuerySingle[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where id=$1", id)
	if err != nil {
		return nil, err
	}
	if session == nil || *session.CreateUser != claims.Subject {
		return nil, errors.New("会话不存在或不属于用户。")
	}
	// 计算消费词元。
	rows, err := storage.RDB.QueryMatrix[struct {
		Model  *string `db:"model"`
		Prompt *int32  `db:"prompt"`
		Answer *int32  `db:"answer"`
	}](ctx, "select model,sum(prompt) as prompt,sum(answer) as answer from consume where session=$1 group by model;", session.ID)
	if err != nil {
		return nil, err
	}
	var llmPrompt, llmAnswer, slmPrompt, slmAnswer, vemPrompt int32
	for _, row := range rows {
		switch *row.Model {
		case "LLM":
			llmPrompt = *row.Prompt
			llmAnswer = *row.Answer
			break
		case "SLM":
			slmPrompt = *row.Prompt
			slmAnswer = *row.Answer
			break
		case "VEM":
			vemPrompt = *row.Prompt
			break
		}
	}
	// 返回消费词元。
	return &consume{LlmPrompt: llmPrompt, LlmAnswer: llmAnswer, SlmPrompt: slmPrompt, SlmAnswer: slmAnswer, VemPrompt: vemPrompt}, nil
}

// session 用户会话。
type session struct {
	// ID 会话编码。
	ID *string `db:"id" json:"id"`
	// Title 会话标题。
	Title *string `db:"title" json:"title"`
	// Summary 前情摘要。
	Summary *string `db:"summary" json:"summary"`
	// Contain 截至序号。
	Contain *int32 `db:"contain" json:"contain"`
	// Current 当前序号。
	Current *int32 `db:"current" json:"current"`
	// CreateUser 创建用户。
	CreateUser *string `db:"create_user" json:"create_user"`
	// CreateTime 创建时间。
	CreateTime *time.Time `db:"create_time" json:"create_time"`
	// UpdateTime 更新时间。
	UpdateTime *time.Time `db:"update_time" json:"update_time"`
}

// message 会话消息。
type message struct {
	Id         *string    `db:"id" json:"id"`
	Session    *string    `db:"session" json:"session"`
	Serial     *int32     `db:"serial" json:"serial"`
	Human      *bool      `db:"human" json:"human"`
	Content    *string    `db:"content" json:"content"`
	Novels     *string    `db:"novels" json:"novels"`
	Segments   *string    `db:"segments" json:"segments"`
	Complete   *bool      `db:"complete" json:"complete"`
	CreateTime *time.Time `db:"create_time" json:"create_time"`
}
