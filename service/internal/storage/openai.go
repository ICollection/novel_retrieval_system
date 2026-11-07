package storage

import (
	"context"
	"fmt"
	"service/internal/config"
	"time"

	openais "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// openai OpenAI 模型连接。
type openai struct {
	client *openais.Client
}

// createOpenaiLLM 创建 OpenAI 大语言模型连接。
func createOpenaiLLM() (*openai, error) {
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Large.Address)}
	if config.Config.Models.Large.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Large.Key))
	}
	client := openais.NewClient(options...)
	return &openai{client: &client}, nil
}

// createOpenaiSLM 创建 OpenAI 小语言模型连接。
func createOpenaiSLM() (*openai, error) {
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Small.Address)}
	if config.Config.Models.Small.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Small.Key))
	}
	client := openais.NewClient(options...)
	return &openai{client: &client}, nil
}

// createOpenaiVMD 创建 OpenAI 向量模型连接。
func createOpenaiVMD() (*openai, error) {
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Embed.Address)}
	if config.Config.Models.Embed.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Embed.Key))
	}
	client := openais.NewClient(options...)
	return &openai{client: &client}, nil
}

// ComputeVector 计算文本向量。
func (openai *openai) ComputeVector(ctx context.Context, text string) ([]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.Models.Embed.Timeout))
	defer cancel()
	// 获取文本向量。
	response, err := openai.client.Embeddings.New(ctx, openais.EmbeddingNewParams{
		Model: config.Config.Models.Embed.Name,
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
