package storage

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"service/internal/config"
	"strings"
	"time"

	openais "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// openai OpenAI 模型连接。
type openai struct {
	client  *openais.Client
	model   string
	timeout int
}

// createOpenaiLLM 创建 OpenAI 大语言模型连接。
func createOpenaiLLM() (*openai, error) {
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Large.Address)}
	if config.Config.Models.Large.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Large.Key))
	}
	client := openais.NewClient(options...)
	return &openai{client: &client, model: config.Config.Models.Large.Name, timeout: config.Config.Models.Large.Timeout}, nil
}

// createOpenaiSLM 创建 OpenAI 小语言模型连接。
func createOpenaiSLM() (*openai, error) {
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Small.Address)}
	if config.Config.Models.Small.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Small.Key))
	}
	client := openais.NewClient(options...)
	return &openai{client: &client, model: config.Config.Models.Small.Name, timeout: config.Config.Models.Small.Timeout}, nil
}

// createOpenaiVMD 创建 OpenAI 向量模型连接。
func createOpenaiVMD() (*openai, error) {
	options := []option.RequestOption{option.WithBaseURL(config.Config.Models.Embed.Address)}
	if config.Config.Models.Embed.Key != "" {
		options = append(options, option.WithAPIKey(config.Config.Models.Embed.Key))
	}
	client := openais.NewClient(options...)
	return &openai{client: &client, model: config.Config.Models.Embed.Name, timeout: config.Config.Models.Embed.Timeout}, nil
}

// ComputeVector 计算文本向量。
func (openai *openai) ComputeVector(ctx context.Context, text string) ([]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.Config.Models.Embed.Timeout))
	defer cancel()
	// 获取文本向量。
	response, err := openai.client.Embeddings.New(ctx, openais.EmbeddingNewParams{
		Model: openai.model,
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

// AnswerDirect 直接回答问题。
func (openai *openai) AnswerDirect(ctx context.Context, system string, prompt string) (*string, error) {
	// 设置请求超时。
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(openai.timeout))
	defer cancel()
	// 调用语言模型。
	response, err := openai.client.Chat.Completions.New(ctx, openais.ChatCompletionNewParams{
		Model: openai.model,
		Messages: []openais.ChatCompletionMessageParamUnion{
			openais.SystemMessage(system),
			openais.UserMessage(prompt),
		},
		Temperature: openais.Float(0),
		MaxTokens:   openais.Int(512),
	})
	if err != nil {
		return nil, fmt.Errorf("调用语言模型失败，详细错误信息：%w", err)
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("调用语言模型失败，详细错误信息：%w", "没有返回结果。")
	}
	// 返回最终答案。
	result := strings.TrimSpace(response.Choices[0].Message.Content)
	return &result, nil
}

// AnswerStream 流式回答问题。
func (openai *openai) AnswerStream(ctx context.Context, system string, prompt string, callback func(kind string, content string) error) (*string, error) {
	// 设置请求超时。
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(openai.timeout))
	defer cancel()
	// 发起流式请求。
	stream := openai.client.Chat.Completions.NewStreaming(ctx, openais.ChatCompletionNewParams{
		Model: openai.model,
		Messages: []openais.ChatCompletionMessageParamUnion{
			openais.SystemMessage(system),
			openais.UserMessage(prompt),
		},
		Temperature: openais.Float(0.7),
		MaxTokens:   openais.Int(4096),
	})
	defer stream.Close()
	// 累积最终答案。
	var answer strings.Builder
	for stream.Next() {
		current := stream.Current()
		if len(current.Choices) == 0 {
			continue
		}
		delta := &current.Choices[0].Delta
		// 思考过程。
		if thinking := getThinking(delta); thinking != nil {
			if err := callback("think", *thinking); err != nil {
				// 思考过程中用户断开。
				result := answer.String()
				return &result, err
			}
		}
		// 答案内容。
		if delta.Content != "" {
			answer.WriteString(delta.Content)
			if err := callback("answer", delta.Content); err != nil {
				// 回答过程中用户断开。
				result := answer.String()
				return &result, err
			}
		}
	}
	// 检查流式错误。
	if err := stream.Err(); err != nil {
		// 模型或网络中断。
		result := answer.String()
		return &result, fmt.Errorf("调用语言模型失败，详细错误信息：%w", err)
	}
	// 返回正常结果。
	result := answer.String()
	return &result, nil
}

// getThinking 获取思考内容。
func getThinking(delta *openais.ChatCompletionChunkChoiceDelta) *string {
	for _, name := range []string{"reasoning_content", "reasoning"} {
		field, exist := delta.JSON.ExtraFields[name]
		if !exist || !field.Valid() {
			continue
		}
		var content string
		if err := json.Unmarshal([]byte(field.Raw()), &content); err != nil {
			continue
		}
		return &content
	}
	return nil
}
