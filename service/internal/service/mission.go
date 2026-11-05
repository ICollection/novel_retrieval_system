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
	result, err := storage.RDB.QueryMatrix[mission](ctx, "select id,novel,status,total_count,embed_count,launch_time,finish_time,update_time,retry_count,message from mission order by update_time desc;")
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Launch 启动分析任务。
func (service *missionService) Launch(ctx *gin.Context, id *string) (*string, error) {
	// 查询分析任务。
	mission, err := storage.RDB.QuerySingle[mission](ctx, "select id,novel,status,total_count,embed_count,launch_time,finish_time,update_time,retry_count,message from mission where id=$1;", id)
	if err != nil {
		return nil, fmt.Errorf("获取分析任务失败，详细错误信息：%w", err)
	}
	// 验证任务状态。
	if *mission.Status != common.MissionStatus01 && *mission.Status != common.MissionStatus07 && *mission.Status != common.MissionStatus08 {
		return nil, errors.New("任务状态错误。")
	}
	// 更新任务状态。
	err = storage.RDB.QueryNaught(ctx, "update mission set status=$1 where id=$2;", common.MissionStatus02, id)
	if err != nil {
		return nil, fmt.Errorf("启动分析任务失败，详细错误信息：%w", err)
	}
	return id, nil
}

// mission 分析任务。
type mission struct {
	Id         *string    `db:"id" json:"id"`
	Novel      *string    `db:"novel" json:"novel"`
	Status     *string    `db:"status" json:"status"`
	TotalCount *int32     `db:"total_count" json:"total_count"`
	EmbedCount *int32     `db:"embed_count" json:"embed_count"`
	LaunchTime *time.Time `db:"launch_time" json:"launch_time,omitempty"`
	FinishTime *time.Time `db:"finish_time" json:"finish_time,omitempty"`
	UpdateTime *time.Time `db:"update_time" json:"update_time"`
	RetryCount *int32     `db:"retry_count" json:"retry_count"`
	Message    *string    `db:"message" json:"message,omitempty"`
}
