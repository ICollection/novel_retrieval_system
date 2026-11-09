package service

import (
	"service/internal/common"
	"service/internal/server"
	"service/internal/storage"
	"strings"
	"time"
	"uuid"

	"github.com/gin-gonic/gin"
)

// converseService 用户对话服务。
type converseService struct{}

// AppendHumanMessage 增加会话用户消息。
func (service *converseService) AppendHumanMessage(ctx *gin.Context, session *string, content *string, novels []string) (*string, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 组织会话消息。
	moment := time.Now().UTC()
	message_id := strings.ToUpper(uuid.NewV4().String())
	// 是否已有会话。
	if session == nil {
		// 组织会话消息。
		session_id := strings.ToUpper(uuid.NewV4().String())
		title := common.Truncate(*content, 50)
		serial := 1
		// 保存会话消息。
		var batch storage.Batch
		batch.Append("insert into session(id,title,summary,contain,current,create_user,create_time,update_time) values($1,$2,$3,$4,$5,$6,$7,$8);", session_id, title, nil, nil, serial, claims.Subject, moment, moment)
		batch.Append("insert into message(id,session,serial,human,content,novels,segments,complete,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", message_id, session_id, serial, true, content, strings.Join(novels, ","), nil, true, moment)
		err := storage.RDB.BatchNaught(ctx, &batch)
		if err != nil {
			return nil, err
		}
		// 返回消息编码。
		return &session_id, nil
	} else {
		// 获取当前序号。
		current, err := storage.RDB.QuerySimple[int](ctx, "select current from session where id=$1", session)
		if err != nil {
			return session, err
		}
		serial := *current + 1
		// 保存会话消息。
		var batch storage.Batch
		batch.Append("update session set current=$1,update_time=$2 where id=$3;", serial, moment, session)
		batch.Append("insert into message(id,session,serial,human,content,novels,segments,complete,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", message_id, session, serial, true, content, strings.Join(novels, ","), nil, true, moment)
		err = storage.RDB.BatchNaught(ctx, &batch)
		if err != nil {
			return nil, err
		}
		// 返回消息编码。
		return session, nil
	}
}

// AppendModelMessage 增加会话模型消息。
func (service *converseService) AppendModelMessage(ctx *gin.Context, session *string, content *string, novels []string, segments []string, complete bool) error {
	// 组织会话消息。
	moment := time.Now().UTC()
	message_id := strings.ToUpper(uuid.NewV4().String())
	current, err := storage.RDB.QuerySimple[int](ctx, "select current from session where id=$1", session)
	if err != nil {
		return err
	}
	serial := *current + 1
	// 保存会话消息。
	var batch storage.Batch
	batch.Append("update session set current=$1,update_time=$2 where id=$3;", serial, moment, session)
	batch.Append("insert into message(id,session,serial,human,content,novels,segments,complete,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", message_id, session, serial, false, content, strings.Join(novels, ","), common.Trim(strings.Join(segments, ",")), complete, moment)
	err = storage.RDB.BatchNaught(ctx, &batch)
	if err != nil {
		return err
	}
	return nil
}

// FilterRecentMessage 查找近期会话消息。
func (service *converseService) FilterRecentMessage(ctx *gin.Context, id *string, recent int32) ([]message, error) {
	// 获取会话记录。
	session, err := storage.RDB.QuerySingle[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where id=$1", id)
	if err != nil {
		return nil, err
	}
	// 获取最近会话。
	messages, err := storage.RDB.QueryMatrix[message](ctx, "select id,session,serial,human,content,novels,segments,complete,create_time from message where session=$1 and serial<=$2 and serial>$3 order by serial;", session.ID, session.Current, *session.Current-recent)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// FilterEffectMessage 查找有效会话消息。
func (service *converseService) FilterEffectMessage(ctx *gin.Context, id string) ([]string, error) {
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
