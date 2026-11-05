package deposit

import (
	"bufio"
	"context"
	"log"
	"os"
	"uuid"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/qdrant/go-client/qdrant"

	"service/internal/config"
)

// Embedding 读取指定文件并写入向量库。
func Embedding(name string) {
	loadFile(name)
}

func loadFile(name string) error {
	// 创建数据连接。
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Embedding.Address)}
	if config.Config.Models.Embedding.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Embedding.Key))
	}
	client1 := openai.NewClient(options...)
	client2, err := qdrant.NewClient(&qdrant.Config{
		Host:   config.Config.Qdrant.Host,
		Port:   config.Config.Qdrant.Port,
		APIKey: config.Config.Qdrant.APIKey,
	})
	if err != nil {
		log.Fatal("创建连接失败，详细错误信息：", err)
	}
	defer client2.Close()
	ctx := context.Background()
	// 打开文件。
	file, err := os.Open(name)
	if err != nil {
		log.Fatal("打开文件失败，详细错误信息：", err)
	}
	defer file.Close()
	// 读取文件。
	scanner := bufio.NewScanner(file)
	text := ""
	for scanner.Scan() {
		line := scanner.Text()
		text += line + "\n"
		if len(text) > 512 {
			handle(&client1, client2, ctx, text, name)
			text = ""
		}
	}
	if len(text) > 0 {
		handle(&client1, client2, ctx, text, name)
		text = ""
	}
	return nil
}
func handle(client1 *openai.Client, client2 *qdrant.Client, ctx context.Context, text string, name string) {
	data := float64To32(getEmbedding(client1, ctx, text))
	upsertPoint(client2, ctx, data, text, name)
}

func getEmbedding(client *openai.Client, ctx context.Context, text string) []float64 {
	response, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: config.Config.Models.Embedding.Name,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
	})
	if err != nil {
		log.Fatalln("生成向量失败，详细错误信息：", err)
	}
	return response.Data[0].Embedding
}

func upsertPoint(client *qdrant.Client, ctx context.Context, data []float32, text string, name string) {
	points := []*qdrant.PointStruct{
		&qdrant.PointStruct{
			Id:      qdrant.NewID(uuid.NewV7().String()),
			Vectors: qdrant.NewVectorsDense(data),
			Payload: qdrant.NewValueMap(map[string]any{"text": text, "name": name}),
		},
	}
	_, err := client.Upsert(ctx, &qdrant.UpsertPoints{CollectionName: config.Config.Qdrant.Collection, Points: points})
	if err != nil {
		log.Fatal("插入向量失败，详细错误信息：", err)
	}
}

// float64To32 浮点64位转32位。
func float64To32(vector []float64) []float32 {
	result := make([]float32, len(vector))
	for i, x := range vector {
		result[i] = float32(x)
	}
	return result
}
