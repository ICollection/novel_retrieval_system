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

// FilterSession 查询用户会话。
func (service *converseService) FilterSession(ctx *gin.Context) ([]session, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 查询用户会话。
	result, err := storage.RDB.QueryMatrix[session](ctx, "select id,title,summary,contain,current,create_user,create_time,update_time from session where create_user=$1 order by update_time desc;", claims.Subject)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AppendHumanMessage 增加会话用户消息。
func (service *converseService) AppendHumanMessage(ctx *gin.Context, session *string, content *string, novels []string) (*string, *string, *int, error) {
	claims := ctx.MustGet("claims").(*server.Claims)
	// 组织会话消息。
	moment := time.Now().UTC()
	message_id := strings.ToUpper(uuid.NewV4().String())
	// 是否已有会话。
	if session == nil {
		// 组织会话消息。
		session_id := strings.ToUpper(uuid.NewV4().String())
		title := (*content)[:50]
		serial := 1
		// 保存会话消息。
		batch := storage.RDB.CreateBatch()
		batch.AppendCommand("insert into session(id,title,summary,contain,current,create_user,create_time,update_time) values($1,$2,$3,$4,$5,$6,$7,$8);", session_id, title, nil, nil, serial, claims.Subject, moment, moment)
		batch.AppendCommand("insert into message(id,session,serial,human,content,novels,segments,complate,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", message_id, session_id, serial, true, content, strings.Join(novels, ","), nil, true, moment)
		err := storage.RDB.BatchNaught(ctx, batch)
		if err != nil {
			return nil, nil, nil, err
		}
		// 返回消息编码。
		return &session_id, &message_id, &serial, nil
	} else {
		// 获取当前序号。
		current, err := storage.RDB.QuerySimple[int](ctx, "select current from session where id=$1", session)
		if err != nil {
			return session, nil, nil, err
		}
		serial := *current + 1
		// 保存会话消息。
		batch := storage.RDB.CreateBatch()
		batch.AppendCommand("update session set current=$1,update_time=$2 where id=$3;", serial, moment, session)
		batch.AppendCommand("insert into message(id,session,serial,human,content,novels,segments,complate,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", message_id, session, serial, true, content, strings.Join(novels, ","), nil, true, moment)
		err = storage.RDB.BatchNaught(ctx, batch)
		if err != nil {
			return nil, nil, nil, err
		}
		// 返回消息编码。
		return session, &message_id, &serial, nil
	}
}

// AppendModelMessage 增加会话模型消息。
func (service *converseService) AppendModelMessage(ctx *gin.Context, session *string, content *string, novels []string, segments []string, complate bool) (*string, *string, *int, error) {
	// 组织会话消息。
	moment := time.Now().UTC()
	message_id := strings.ToUpper(uuid.NewV4().String())
	current, err := storage.RDB.QuerySimple[int](ctx, "select current from session where id=$1", session)
	if err != nil {
		return session, nil, nil, err
	}
	serial := *current + 1
	// 保存会话消息。
	batch := storage.RDB.CreateBatch()
	batch.AppendCommand("update session set current=$1,update_time=$2 where id=$3;", serial, moment, session)
	batch.AppendCommand("insert into message(id,session,serial,human,content,novels,segments,complate,create_time) values($1,$2,$3,$4,$5,$6,$7,$8,$9);", message_id, session, serial, false, content, strings.Join(novels, ","), common.Trim(strings.Join(segments, ",")), complate, moment)
	err = storage.RDB.BatchNaught(ctx, batch)
	if err != nil {
		return nil, nil, nil, err
	}
	// 返回消息编码。
	return session, &message_id, &serial, nil
}

// session 用户会话。
type session struct {
	Id         *string    `db:"id"`
	Title      *string    `db:"title"`
	Summary    *string    `db:"summary"`
	Contain    *string    `db:"contain"`
	Current    *int32     `db:"current"`
	CreateUser *string    `db:"create_user"`
	CreateTime *time.Time `db:"create_time"`
	UpdateTime *time.Time `db:"update_time"`
}
