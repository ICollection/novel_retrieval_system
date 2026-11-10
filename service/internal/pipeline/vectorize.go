package pipeline

import (
	"context"
	"errors"
	"service/internal/common"
	"service/internal/config"
	"service/internal/storage"
	"strings"
	"time"
)

// vectorize 切片向量入库。
//
// param: ctx 上下文。
// param: mission 任务编码。
// param: novel 小说编码。
// return: 已入库的分片序号，错误信息。
func vectorize(ctx context.Context, mission string, novel string) (int32, error) {
	// 获取起止向量。
	total_count, err := storage.RDB.QuerySimple[int32](ctx, "select total_count from mission where id=$1", mission)
	if err != nil {
		return 0, err
	}
	embed_count, err := storage.RDB.QuerySimple[int32](ctx, "select embed_count from mission where id=$1", mission)
	if err != nil {
		return 0, err
	}
	// 处理剩余切片。
	current := *embed_count
	for i := *embed_count; i < *total_count; i += 20 {
		// 验证是否退出。
		if ctx.Err() != nil {
			return current, ctx.Err()
		}
		// 获取一段切片。
		segments, err := storage.RDB.QueryMatrix[segment](ctx, "select id,chapter,serial,prefix,middle,suffix from segment where novel=$1 and serial>$2 and serial<=$3 order by serial;", novel, i, i+20)
		if err != nil {
			return current, err
		}
		// 依次保存切片。
		for _, segment := range segments {
			content := segment.getContent()
			original, _, err := storage.VEM.ComputeVector(ctx, content)
			if err != nil {
				return current, err
			}
			var abstract []float32
			if config.Cfg.Station.Abstract == true {
				summary, _, _, err := storage.SLM.AnswerDirect(ctx, config.Abstract, "======== 以下为待提炼文本 ========\n"+content, 8192)
				if err != nil {
					return current, err
				}
				if summary == nil {
					return current, errors.New("提取摘要为空。")
				}
				abstract, _, err = storage.VEM.ComputeVector(ctx, *summary)
				if err != nil {
					return current, err
				}
			}
			matching1, matching2 := common.GetMatching(content)
			err = storage.VDB.UpsertPoints(ctx, *segment.Segment, original, abstract, matching1, matching2, map[string]any{"novel": novel, "chapter": *segment.Chapter, "serial": *segment.Serial})
			if err != nil {
				return current, err
			}
			current = *segment.Serial
		}
		// 更新处理进度。
		err = storage.RDB.QueryNaught(ctx, "update mission set embed_count=$1,update_time=$2 where id=$3;", current, time.Now().UTC(), mission)
		if err != nil {
			return current, err
		}
	}
	return current, nil
}

// segment 文章分片。
type segment struct {
	Segment *string `db:"id"`
	Chapter *string `db:"chapter"`
	Serial  *int32  `db:"serial"`
	Prefix  *string `db:"prefix"`
	Middle  *string `db:"middle"`
	Suffix  *string `db:"suffix"`
}

// getContent 获取分片内容。
//
// return: 由前缀、正文、后缀拼接的分片内容。
func (segment *segment) getContent() string {
	builder := strings.Builder{}
	if segment.Prefix != nil {
		builder.WriteString(*segment.Prefix)
		builder.WriteString("\n")
	}
	builder.WriteString(*segment.Middle)
	if segment.Suffix != nil {
		builder.WriteString("\n")
		builder.WriteString(*segment.Suffix)
	}
	return builder.String()
}
