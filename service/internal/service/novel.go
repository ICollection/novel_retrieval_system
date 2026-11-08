package service

import (
	"errors"
	"service/internal/common"
	"service/internal/storage"
	"time"

	"github.com/gin-gonic/gin"
)

// novelService 小说对象服务。
type novelService struct {
}

// Filter 查询小说对象。
func (service *novelService) Filter(ctx *gin.Context) ([]novel, error) {
	result, err := storage.RDB.QueryMatrix[novel](ctx, "select novel.id,novel.origin,novel.title,mission.id as mission,mission.status,novel.character_count,novel.chapter_count,account.id as create_code,account.name as create_name,novel.create_time,novel.update_time from novel inner join account on account.id=novel.create_user inner join mission on mission.novel=novel.id order by novel.create_time;")
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Modify 修改小说对象。
func (service *novelService) Modify(ctx *gin.Context, id *string, title *string) (*string, error) {
	err := storage.RDB.QueryNaught(ctx, "update novel set title=$1,update_time=$2 where id=$3;", title, time.Now().UTC(), id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// Delete 删除小说对象。
func (service *novelService) Delete(ctx *gin.Context, id *string) (*string, error) {
	// 验证小说状态。
	novel, err := storage.RDB.QuerySingle[novel](ctx, "select novel.id,novel.origin,novel.title,mission.id as mission,mission.status,novel.character_count,novel.chapter_count,account.id as create_code,account.name as create_name,novel.create_time,novel.update_time from novel inner join account on account.id=novel.create_user inner join mission on mission.novel=novel.id where novel.id=$1", id)
	if err != nil {
		return nil, err
	}
	if *novel.Status != common.MissionStatus01 && *novel.Status != common.MissionStatus06 && *novel.Status != common.MissionStatus07 && *novel.Status != common.MissionStatus08 {
		return nil, errors.New("任务状态错误。")
	}
	kind, err := storage.RDB.QuerySimple[string](ctx, "select type from origin where id=$1", novel.Origin)
	if err != nil {
		return nil, err
	}
	// 删除关系数据。
	var batch storage.Batch
	batch.Append("delete from segment where novel=$1", novel.Id)
	batch.Append("delete from chapter where novel=$1", novel.Id)
	batch.Append("delete from mission where id=$1", novel.Mission)
	batch.Append("delete from novel where id=$1", novel.Id)
	batch.Append("delete from origin where id=$1", novel.Origin)
	err = storage.RDB.BatchNaught(ctx, &batch)
	if err != nil {
		return nil, err
	}
	// 删除向量数据。
	err = storage.VDB.DeletePoints(ctx, "novel", *novel.Id)
	if err != nil {
		return nil, err
	}
	// 删除归一文件。
	storage.OSS.Delete(ctx, "Normal", *novel.Origin+".md")
	// 删除原始文件。
	storage.OSS.Delete(ctx, "Origin", *novel.Origin+"."+*kind)
	// 返回小说编码。
	return novel.Id, nil
}

// novel 小说对象。
type novel struct {
	Id             *string    `db:"id" json:"id"`
	Origin         *string    `db:"origin" json:"origin"`
	Title          *string    `db:"title" json:"title"`
	Mission        *string    `db:"mission" json:"mission"`
	Status         *string    `db:"status" json:"status"`
	CharacterCount *int32     `db:"character_count" json:"character_count"`
	ChapterCount   *int32     `db:"chapter_count" json:"chapter_count"`
	CreateCode     *string    `db:"create_code" json:"create_code"`
	CreateName     *string    `db:"create_name" json:"create_name"`
	CreateTime     *time.Time `db:"create_time" json:"create_time"`
	UpdateTime     *time.Time `db:"update_time" json:"update_time"`
}
