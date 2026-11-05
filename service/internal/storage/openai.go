package storage

import (
	"context"
	"fmt"
	"service/internal/config"

	openais "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// openai OpenAI 模型连接。
type openai struct {
	client *openais.Client
}

// createOpenaiLLM 创建 OpenAI 语言模型连接。
func createOpenaiLLM() (*openai, error) {
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Language.Address)}
	if config.Config.Models.Language.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Language.Key))
	}
	client := openais.NewClient(options...)
	return &openai{client: &client}, nil
}

// createOpenaiVMD 创建 OpenAI 向量模型连接。
func createOpenaiVMD() (*openai, error) {
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Embedding.Address)}
	if config.Config.Models.Embedding.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Embedding.Key))
	}
	client := openais.NewClient(options...)
	return &openai{client: &client}, nil
}

// GetEmbedding 获取文本向量。
func (openai *openai) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	// 获取文本向量。
	response, err := openai.client.Embeddings.New(ctx, openais.EmbeddingNewParams{
		Model: config.Config.Models.Embedding.Name,
		Input: openais.EmbeddingNewParamsInputUnion{
			OfString: openais.String(text),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("获取文本向量失败，详细错误信息： %w", err)
	}
	// 转换向量类型。
	result := make([]float32, len(response.Data[0].Embedding))
	for i, x := range response.Data[0].Embedding {
		result[i] = float32(x)
	}
	return result, nil
}
