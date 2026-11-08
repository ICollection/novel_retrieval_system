package storage

import (
	"context"
	"log"
	"service/internal/config"

	qdrants "github.com/qdrant/go-client/qdrant"
)

// qdrant Qdrant 向量数据库连接。
type qdrant struct {
	client *qdrants.Client
}

// createQdrant 创建 Qdrant 向量数据库连接。
//
// return: Qdrant 向量数据库连接，错误信息。
func createQdrant() (*qdrant, error) {
	client, err := qdrants.NewClient(&qdrants.Config{Host: config.Config.Qdrant.Host, Port: config.Config.Qdrant.Port, APIKey: config.Config.Qdrant.APIKey})
	if err != nil {
		log.Println("[VDB]创建连接失败，详细错误信息：", err.Error())
		return nil, err
	}
	return &qdrant{client: client}, nil
}

// CreateCollection 创建向量集合。
//
// return: 错误信息。
func (qdrant *qdrant) CreateCollection() error {
	ctx := context.Background()
	if exists, err := qdrant.client.CollectionExists(ctx, config.Config.Qdrant.Collection); err != nil {
		log.Println("[VDB]检测集合失败，详细错误信息：", err.Error())
		return err
	} else if exists == false {
		request := qdrants.CreateCollection{
			CollectionName: config.Config.Qdrant.Collection,
			VectorsConfig: qdrants.NewVectorsConfig(&qdrants.VectorParams{
				Size:     1024,
				Distance: qdrants.Distance_Cosine,
				Datatype: qdrants.Datatype_Float16.Enum(),
			}),
		}
		if err = qdrant.client.CreateCollection(ctx, &request); err != nil {
			log.Println("[VDB]创建集合失败，详细错误信息：", err.Error())
			return err
		}
	}
	return nil
}

// UpsertPoints 保存向量数据。
//
// param: ctx 上下文。
// param: id 向量标识。
// param: vector 向量数据。
// param: payload 向量载荷。
// return: 错误信息。
func (qdrant *qdrant) UpsertPoints(ctx context.Context, id string, vector []float32, payload map[string]any) error {
	points := []*qdrants.PointStruct{&qdrants.PointStruct{Id: qdrants.NewID(id), Vectors: qdrants.NewVectorsDense(vector), Payload: qdrants.NewValueMap(payload)}}
	_, err := qdrant.client.Upsert(ctx, &qdrants.UpsertPoints{CollectionName: config.Config.Qdrant.Collection, Points: points})
	return err
}

// DeletePoints 条件删除向量。
//
// param: ctx 上下文。
// param: field 条件字段。
// param: keyword 条件键值。
// return: 错误信息。
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

// FilterPoints 查询召回分片。
//
// param: ctx 上下文。
// param: vector 查询向量。
// param: novels 小说范围。
// param: limit 查询数量。
// param: threshold 相似度值。
// return: 召回分片，错误信息。
func (qdrant *qdrant) FilterPoints(ctx context.Context, vector []float32, novels []string, limit int32, threshold float32) ([]RecallSegment, error) {
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
	segments := make([]RecallSegment, 0, len(points))
	for _, point := range points {
		segments = append(segments, RecallSegment{
			Id:      point.Id.GetUuid(),
			Score:   point.Score,
			Novel:   point.Payload["novel"].GetStringValue(),
			Chapter: point.Payload["chapter"].GetStringValue(),
			Serial:  int32(point.Payload["serial"].GetIntegerValue()),
		})
	}
	return segments, nil
}

// RecallSegment 召回分片。
type RecallSegment struct {
	Id      string
	Novel   string
	Chapter string
	Serial  int32
	Score   float32
}
