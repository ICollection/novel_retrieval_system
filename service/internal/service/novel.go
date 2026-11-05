package service

import (
	"service/internal/storage"
	"time"

	"github.com/gin-gonic/gin"
)

// novelService 小说对象服务。
type novelService struct {
}

// Filter 查询小说对象。
func (service *novelService) Filter(ctx *gin.Context) ([]novel, error) {
	result, err := storage.RDB.QueryMatrix[novel](ctx, "select novel.id,novel.origin,novel.title,mission.status,novel.character_count,novel.chapter_count,account.id as create_code,account.name as create_name,novel.create_time,novel.update_time from novel inner join account on account.id=novel.create_user inner join mission on mission.novel=novel.id order by novel.create_time;")
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Modify 修改小说对象。
func (service *novelService) Modify(ctx *gin.Context, id *string, title *string) (*string, error) {
	err := storage.RDB.QueryNaught(ctx, "update novel set title=$1,update_time=$2 where id=$3;", title, time.Now(), id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// novel 小说对象。
type novel struct {
	Id             *string    `db:"id" json:"id"`
	Origin         *string    `db:"origin" json:"origin"`
	Title          *string    `db:"title" json:"title"`
	Status         *string    `db:"status" json:"status"`
	CharacterCount *int32     `db:"character_count" json:"character_count"`
	ChapterCount   *int32     `db:"chapter_count" json:"chapter_count"`
	CreateCode     *string    `db:"create_code" json:"create_code"`
	CreateName     *string    `db:"create_name" json:"create_name"`
	CreateTime     *time.Time `db:"create_time" json:"create_time"`
	UpdateTime     *time.Time `db:"update_time" json:"update_time"`
}
