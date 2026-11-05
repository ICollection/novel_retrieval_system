package service

import (
	"errors"
	"fmt"
	"service/internal/common"
	"service/internal/storage"
	"time"

	"github.com/gin-gonic/gin"
)

// missionService 分析任务服务。
type missionService struct {
}

// Filter 查询分析任务。
func (service *missionService) Filter(ctx *gin.Context) ([]mission, error) {
	result, err := storage.RDB.QueryMatrix[mission](ctx, "select mission.id,novel.id as novel_code,novel.title as novel_name,mission.status,mission.total_count,mission.embed_count,mission.launch_time,mission.finish_time,mission.update_time,mission.retry_count,mission.message from mission inner join novel on novel.id=mission.novel order by mission.update_time desc;")
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Launch 启动分析任务。
func (service *missionService) Launch(ctx *gin.Context, id *string) (*string, error) {
	// 查询分析任务。
	mission, err := storage.RDB.QuerySingle[mission](ctx, "select mission.id,novel.id as novel_code,novel.title as novel_name,mission.status,mission.total_count,mission.embed_count,mission.launch_time,mission.finish_time,mission.update_time,mission.retry_count,mission.message from mission inner join novel on novel.id=mission.novel where mission.id=$1;", id)
	if err != nil {
		return nil, fmt.Errorf("获取分析任务失败，详细错误信息：%w", err)
	}
	// 验证任务状态。
	if *mission.Status != common.MissionStatus01 && *mission.Status != common.MissionStatus07 && *mission.Status != common.MissionStatus08 {
		return nil, errors.New("任务状态错误。")
	}
	// 更新任务状态。
	err = storage.RDB.QueryNaught(ctx, "update mission set status=$1,launch_time=$2 where id=$3;", common.MissionStatus02, time.Now().UTC(), id)
	if err != nil {
		return nil, fmt.Errorf("启动分析任务失败，详细错误信息：%w", err)
	}
	return id, nil
}

// Cancel 取消分析任务。
func (service *missionService) Cancel(ctx *gin.Context, id *string) (*string, error) {
	// 查询分析任务。
	mission, err := storage.RDB.QuerySingle[mission](ctx, "select mission.id,novel.id as novel_code,novel.title as novel_name,mission.status,mission.total_count,mission.embed_count,mission.launch_time,mission.finish_time,mission.update_time,mission.retry_count,mission.message from mission inner join novel on novel.id=mission.novel where mission.id=$1;", id)
	if err != nil {
		return nil, fmt.Errorf("获取分析任务失败，详细错误信息：%w", err)
	}
	// 验证任务状态。
	if *mission.Status != common.MissionStatus02 && *mission.Status != common.MissionStatus03 && *mission.Status != common.MissionStatus04 && *mission.Status != common.MissionStatus05 && *mission.Status != common.MissionStatus06 {
		return nil, errors.New("任务状态错误。")
	}
	// 更新任务状态。
	err = storage.RDB.QueryNaught(ctx, "update mission set status=$1 where id=$2;", common.MissionStatus08, id)
	if err != nil {
		return nil, fmt.Errorf("取消分析任务失败，详细错误信息：%w", err)
	}
	return id, nil
}

// mission 分析任务。
type mission struct {
	Id         *string    `db:"id" json:"id"`
	NovelCode  *string    `db:"novel_code" json:"novel_code"`
	NovelName  *string    `db:"novel_name" json:"novel_name"`
	Status     *string    `db:"status" json:"status"`
	TotalCount *int32     `db:"total_count" json:"total_count"`
	EmbedCount *int32     `db:"embed_count" json:"embed_count"`
	LaunchTime *time.Time `db:"launch_time" json:"launch_time,omitempty"`
	FinishTime *time.Time `db:"finish_time" json:"finish_time,omitempty"`
	UpdateTime *time.Time `db:"update_time" json:"update_time"`
	RetryCount *int32     `db:"retry_count" json:"retry_count"`
	Message    *string    `db:"message" json:"message,omitempty"`
}
