package storage

import (
	"context"
	"log/slog"
	"service/internal/config"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantClient Qdrant 向量数据库连接。
type QdrantClient struct {
	client     *qdrant.Client
	collection string
}

// newQdrantClient 创建 Qdrant 向量数据库连接。
//
// param: cfg 数据库连接配置。
// return: Qdrant 向量数据库连接，错误信息。
func newQdrantClient(cfg config.Qdrant) (*QdrantClient, error) {
	client, err := qdrant.NewClient(&qdrant.Config{Host: cfg.Host, Port: cfg.Port, APIKey: cfg.APIKey})
	if err != nil {
		slog.Error("[VDB]创建连接失败", "error", err)
		return nil, err
	}
	return &QdrantClient{client: client, collection: cfg.Collection}, nil
}

// CreateCollection 创建向量集合。
//
// return: 错误信息。
func (qc *QdrantClient) CreateCollection() error {
	ctx := context.Background()
	if exists, err := qc.client.CollectionExists(ctx, qc.collection); err != nil {
		slog.Error("[VDB]检测集合失败", "error", err)
		return err
	} else if exists == false {
		request := qdrant.CreateCollection{
			CollectionName: qc.collection,
			VectorsConfig: qdrant.NewVectorsConfigMap(map[string]*qdrant.VectorParams{
				"original": {
					Size:     1024,
					Distance: qdrant.Distance_Cosine,
					Datatype: qdrant.Datatype_Float16.Enum(),
				},
				"abstract": {
					Size:     1024,
					Distance: qdrant.Distance_Cosine,
					Datatype: qdrant.Datatype_Float16.Enum(),
				},
			}),
			SparseVectorsConfig: &qdrant.SparseVectorConfig{
				Map: map[string]*qdrant.SparseVectorParams{
					"matching": {
						Modifier: qdrant.Modifier_Idf.Enum(),
					},
				},
			},
		}
		if err = qc.client.CreateCollection(ctx, &request); err != nil {
			slog.Error("[VDB]创建集合失败", "error", err)
			return err
		}
	}
	return nil
}

// FilterPoints 查询向量数据。
//
// param: ctx 上下文。
// param: question 问题向量。
// param: matching1 稀疏向量。
// param: matching2 稀疏向量。
// param: novels 小说范围。
// param: limit 查询数量。
// param: threshold 相似度值。
// return: 向量数据，错误信息。
func (qc *QdrantClient) FilterPoints(ctx context.Context, question []float32, matching1 []uint32, matching2 []float32, novels []string, limit int, threshold float32) ([]*qdrant.ScoredPoint, error) {
	var (
		points []*qdrant.ScoredPoint
		err    error
	)
	filter := &qdrant.Filter{Must: []*qdrant.Condition{qdrant.NewMatchKeywords("novel", novels...)}}
	if matching1 == nil || matching2 == nil || len(matching1) == 0 || len(matching1) != len(matching2) {
		points, err = qc.client.Query(ctx, &qdrant.QueryPoints{
			CollectionName: qc.collection,
			Prefetch: []*qdrant.PrefetchQuery{
				{
					Query:          qdrant.NewQueryNearest(qdrant.NewVectorInputDense(question)),
					Using:          new("original"),
					Filter:         filter,
					Limit:          new(uint64(limit)),
					ScoreThreshold: new(threshold),
				},
				{
					Query:          qdrant.NewQueryNearest(qdrant.NewVectorInputDense(question)),
					Using:          new("abstract"),
					Filter:         filter,
					Limit:          new(uint64(limit)),
					ScoreThreshold: new(threshold),
				},
			},
			Query:       qdrant.NewQueryFusion(qdrant.Fusion_RRF),
			Filter:      filter,
			Limit:       new(uint64(limit)),
			WithPayload: qdrant.NewWithPayload(true),
		})
	} else {
		points, err = qc.client.Query(ctx, &qdrant.QueryPoints{
			CollectionName: qc.collection,
			Prefetch: []*qdrant.PrefetchQuery{
				{
					Query:          qdrant.NewQueryNearest(qdrant.NewVectorInputDense(question)),
					Using:          new("original"),
					Filter:         filter,
					Limit:          new(uint64(limit)),
					ScoreThreshold: new(threshold),
				},
				{
					Query:          qdrant.NewQueryNearest(qdrant.NewVectorInputDense(question)),
					Using:          new("abstract"),
					Filter:         filter,
					Limit:          new(uint64(limit)),
					ScoreThreshold: new(threshold),
				},
				{
					Query:  qdrant.NewQueryNearest(qdrant.NewVectorInputSparse(matching1, matching2)),
					Using:  new("matching"),
					Filter: filter,
					Limit:  new(uint64(limit)),
				},
			},
			Query:       qdrant.NewQueryFusion(qdrant.Fusion_RRF),
			Filter:      filter,
			Limit:       new(uint64(limit)),
			WithPayload: qdrant.NewWithPayload(true),
		})
	}
	if err != nil {
		slog.Error("[VDB]查询向量失败", "error", err)
		return nil, err
	}
	return points, nil
}

// UpsertPoints 保存向量数据。
//
// param: ctx 上下文。
// param: id 向量标识。
// param: original 原文向量。
// param: abstract 摘要向量。
// param: matching1 稀疏向量。
// param: matching2 稀疏向量。
// param: payload 向量载荷。
// return: 错误信息。
func (qc *QdrantClient) UpsertPoints(ctx context.Context, id string, original []float32, abstract []float32, matching1 []uint32, matching2 []float32, payload map[string]any) error {
	vectors := map[string]*qdrant.Vector{}
	vectors["original"] = qdrant.NewVectorDense(original)
	if abstract != nil {
		vectors["abstract"] = qdrant.NewVectorDense(abstract)
	}
	if matching1 != nil && matching2 != nil && len(matching1) > 0 && len(matching2) > 0 {
		vectors["matching"] = qdrant.NewVectorSparse(matching1, matching2)
	}
	points := []*qdrant.PointStruct{
		&qdrant.PointStruct{
			Id:      qdrant.NewID(id),
			Vectors: qdrant.NewVectorsMap(vectors),
			Payload: qdrant.NewValueMap(payload),
		},
	}
	_, err := qc.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: qc.collection,
		Points:         points,
	})
	if err != nil {
		slog.Error("[VDB]保存向量失败", "error", err)
		return err
	}
	return nil
}

// DeletePoints 条件删除向量。
//
// param: ctx 上下文。
// param: field 条件字段。
// param: keyword 条件键值。
// return: 错误信息。
func (qc *QdrantClient) DeletePoints(ctx context.Context, field string, keyword string) error {
	_, err := qc.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: qc.collection,
		Points: qdrant.NewPointsSelectorFilter(&qdrant.Filter{
			Must: []*qdrant.Condition{qdrant.NewMatchKeyword(field, keyword)},
		}),
		Wait: new(true),
	})
	if err != nil {
		slog.Error("[VDB]删除向量失败", "error", err)
		return err
	}
	return nil
}
