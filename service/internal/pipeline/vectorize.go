package pipeline

import (
	"context"
	"fmt"
	"service/internal/storage"
	"time"
)

// vectorize 切片向量入库。
func vectorize(ctx context.Context, mission string, novel string) (int32, error) {
	// 获取起止向量。
	total_count, err := storage.RDB.QuerySimple[int32](ctx, "select total_count from mission where id=$1", mission)
	if err != nil {
		return 0, fmt.Errorf("获取任务状态失败，详细错误信息：%w", err)
	}
	embed_count, err := storage.RDB.QuerySimple[int32](ctx, "select embed_count from mission where id=$1", mission)
	if err != nil {
		return 0, fmt.Errorf("获取任务状态失败，详细错误信息：%w", err)
	}
	// 处理剩余切片。
	current := *embed_count
	for i := *embed_count; i < *total_count; i += 100 {
		// 验证是否退出。
		if ctx.Err() != nil {
			return current, ctx.Err()
		}
		// 获取一段切片。
		segments, err := storage.RDB.QueryMatrix[segment](ctx, "select id,chapter,serial,content from segment where novel=$1 and serial>$2 and serial<=$3 order by serial;", novel, i, i+100)
		if err != nil {
			return current, fmt.Errorf("切片向量入库失败，详细错误信息：%w", err)
		}
		// 依次保存切片。
		for _, segment := range segments {
			vector, err := storage.VMD.GetEmbedding(ctx, *segment.Content)
			if err != nil {
				return current, fmt.Errorf("切片向量入库失败，详细错误信息：%w", err)
			}
			err = storage.VDB.UpsertPoints(ctx, *segment.Segment, vector, map[string]any{"novel": novel, "chapter": *segment.Chapter, "serial": *segment.Serial})
			if err != nil {
				return current, fmt.Errorf("切片向量入库失败，详细错误信息：%w", err)
			}
			current = *segment.Serial
		}
		// 更新处理进度。
		err = storage.RDB.QueryNaught(ctx, "update mission set embed_count=$1,update_time=$2 where id=$3;", current, time.Now().UTC(), mission)
		if err != nil {
			return current, fmt.Errorf("切片向量入库失败，详细错误信息：%w", err)
		}
	}
	return current, nil
}

// segment 文章分片。
type segment struct {
	Segment *string `db:"id"`
	Chapter *string `db:"chapter"`
	Serial  *int32  `db:"serial"`
	Content *string `db:"content"`
}
