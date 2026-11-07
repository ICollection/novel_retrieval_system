package storage

import (
	"context"
	"fmt"
	"service/internal/config"

	qdrants "github.com/qdrant/go-client/qdrant"
)

// qdrant Qdrant 向量数据库连接。
type qdrant struct {
	client *qdrants.Client
}

// createQdrant 创建 Qdrant 向量数据库连接。
func createQdrant() (*qdrant, error) {
	client, err := qdrants.NewClient(&qdrants.Config{Host: config.Config.Qdrant.Host, Port: config.Config.Qdrant.Port, APIKey: config.Config.Qdrant.APIKey})
	if err != nil {
		return nil, fmt.Errorf("创建 Qdrant 向量数据库连接失败，详细错误信息：%w", err)
	}
	return &qdrant{client: client}, nil
}

// UpsertPoints 保存向量数据。
func (qdrant *qdrant) UpsertPoints(ctx context.Context, id string, vector []float32, payload map[string]any) error {
	points := []*qdrants.PointStruct{&qdrants.PointStruct{Id: qdrants.NewID(id), Vectors: qdrants.NewVectorsDense(vector), Payload: qdrants.NewValueMap(payload)}}
	_, err := qdrant.client.Upsert(ctx, &qdrants.UpsertPoints{CollectionName: config.Config.Qdrant.Collection, Points: points})
	return err
}

// DeletePoints 条件删除向量。
func (qdrant *qdrant) DeletePoints(ctx context.Context, field string, keyword string) error {
	_, err := qdrant.client.Delete(ctx, &qdrants.DeletePoints{
		CollectionName: config.Config.Qdrant.Collection,
		Points: qdrants.NewPointsSelectorFilter(&qdrants.Filter{
			Must: []*qdrants.Condition{
				qdrants.NewMatchKeyword(field, keyword),
			},
		}),
		Wait: new(true),
	})
	return err
}

// FilterPoints 条件查询向量。
func (qdrant *qdrant) FilterPoints(ctx context.Context, vector []float32, novels []string, limit int32, threshold float32) ([]segment, error) {
	// 查询向量数据。
	points, err := qdrant.client.Query(ctx, &qdrants.QueryPoints{
		CollectionName: config.Config.Qdrant.Collection,
		Query:          qdrants.NewQueryNearest(qdrants.NewVectorInputDense(vector)),
		Filter:         &qdrants.Filter{Must: []*qdrants.Condition{qdrants.NewMatchKeywords("novel", novels...)}},
		Limit:          new(uint64(limit)),
		ScoreThreshold: new(threshold),
		WithPayload:    qdrants.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}
	// 整理选出切片。
	segments := make([]segment, 0, len(points))
	for _, point := range points {
		segments = append(segments, segment{
			Id:      point.Id.GetUuid(),
			Score:   point.Score,
			Novel:   point.Payload["novel"].GetStringValue(),
			Chapter: point.Payload["chapter"].GetStringValue(),
			Serial:  int32(point.Payload["serial"].GetIntegerValue()),
		})
	}
	return segments, nil
}

// 选出切片。
type segment struct {
	Id      string
	Novel   string
	Chapter string
	Serial  int32
	Score   float32
}
