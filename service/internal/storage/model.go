package storage

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"service/internal/common"
	"service/internal/config"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// ModelClient 模型服务连接。
type ModelClient struct {
	client  *openai.Client
	model   string
	sign    string
	timeout int
}

// newModelClient 创建模型服务连接。
//
// param: cfg 模型配置。
// param: sign 模型标识。
// return: 模型服务连接，错误信息。
func newModelClient(cfg config.Model, sign string) (*ModelClient, error) {
	options := []option.RequestOption{option.WithBaseURL(cfg.Address)}
	if cfg.Key != "" {
		options = append(options, option.WithAPIKey(cfg.Key))
	}
	client := openai.NewClient(options...)
	return &ModelClient{client: &client, model: cfg.Name, sign: sign, timeout: cfg.Timeout}, nil
}

// ComputeVector 计算文本向量。
//
// param: ctx 上下文。
// param: text 待算文本。
// return: 文本向量，输入词元，错误信息。
func (mc *ModelClient) ComputeVector(ctx context.Context, text string) ([]float32, int32, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(mc.timeout))
	defer cancel()
	// 获取文本向量。
	response, err := mc.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: mc.model,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
	})
	if err != nil {
		slog.Error(mc.sign+"计算向量失败", "error", err)
		return nil, 0, err
	}
	// 转换向量类型。
	result := make([]float32, len(response.Data[0].Embedding))
	for i, x := range response.Data[0].Embedding {
		result[i] = float32(x)
	}
	slog.Info(mc.sign+"模型完成向量", "tokens", response.Usage.PromptTokens)
	return result, int32(response.Usage.PromptTokens), nil
}

// AnswerDirect 直接回答问题。
//
// param: ctx 上下文。
// param: system 系统提示词。
// param: prompt 用户提示词。
// param: tokens 词元限制。
// return: 答案文本，输入词元，输出词元，错误信息。
func (mc *ModelClient) AnswerDirect(ctx context.Context, system string, prompt string, tokens int32) (*string, int32, int32, error) {
	slog.Debug(mc.sign+"调用模型提问", "system", system, "prompt", prompt)
	// 设置请求超时。
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(mc.timeout))
	defer cancel()
	// 调用语言模型。
	response, err := mc.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: mc.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(prompt),
		},
		Temperature: openai.Float(0),
		MaxTokens:   openai.Int(int64(tokens)),
	}, option.WithJSONSet("chat_template_kwargs", map[string]any{"enable_thinking": false}))
	if err != nil {
		slog.Error(mc.sign+"生成回答失败", "error", err)
		return nil, 0, 0, err
	}
	if len(response.Choices) == 0 {
		err := errors.New("没有返回结果。")
		slog.Error(mc.sign+"生成回答失败", "error", err)
		return nil, int32(response.Usage.PromptTokens), int32(response.Usage.CompletionTokens), err
	}
	// 返回最终答案。
	slog.Info(mc.sign+"模型完成回答", "token1", response.Usage.PromptTokens, "token2", response.Usage.CompletionTokens, "answer", response.Choices[0].Message.Content)
	result := common.Trim(response.Choices[0].Message.Content)
	return result, int32(response.Usage.PromptTokens), int32(response.Usage.CompletionTokens), nil
}

// AnswerStream 流式回答问题。
//
// param: ctx 上下文。
// param: system 系统提示词。
// param: prompt 用户提示词。
// param: callback 回调函数。
// return: 答案文本，输入词元，输出词元，错误信息。
func (mc *ModelClient) AnswerStream(ctx context.Context, system string, prompt string, callback func(kind string, content any) error) (*string, int32, int32, error) {
	slog.Debug(mc.sign+"调用模型提问", "system", system, "prompt", prompt)
	// 设置请求超时。
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(mc.timeout))
	defer cancel()
	// 发起流式请求。
	stream := mc.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: mc.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(prompt),
		},
		Temperature: openai.Float(0.7),
		MaxTokens:   openai.Int(131072),
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		},
	})
	defer stream.Close()
	// 累积最终答案。
	var answer strings.Builder
	var usage openai.CompletionUsage
	for stream.Next() {
		current := stream.Current()
		if current.Usage.TotalTokens > 0 {
			usage = current.Usage
		}
		if len(current.Choices) == 0 {
			continue
		}
		delta := &current.Choices[0].Delta
		if thinking := getThinking(delta); thinking != nil {
			if err := callback("think", *thinking); err != nil {
				// 思考过程中用户断开。
				slog.Error(mc.sign+"用户中断思考", "token1", usage.PromptTokens, "token2", usage.CompletionTokens, "error", err)
				result := common.Trim(answer.String())
				return result, int32(usage.PromptTokens), int32(usage.CompletionTokens), err
			}
		}
		// 答案内容。
		if delta.Content != "" {
			answer.WriteString(delta.Content)
			if err := callback("answer", delta.Content); err != nil {
				// 回答过程中用户断开。
				slog.Error(mc.sign+"用户中断回答", "token1", usage.PromptTokens, "token2", usage.CompletionTokens, "error", err)
				result := common.Trim(answer.String())
				return result, int32(usage.PromptTokens), int32(usage.CompletionTokens), err
			}
		}
	}
	// 检查流式错误。
	if err := stream.Err(); err != nil {
		// 模型或网络中断。
		slog.Error(mc.sign+"模型中断回答", "token1", usage.PromptTokens, "token2", usage.CompletionTokens, "error", err)
		result := common.Trim(answer.String())
		return result, int32(usage.PromptTokens), int32(usage.CompletionTokens), err
	}
	// 返回正常结果。
	slog.Info(mc.sign+"模型完成回答", "token1", usage.PromptTokens, "token2", usage.CompletionTokens, "answer", answer.String())
	result := common.Trim(answer.String())
	return result, int32(usage.PromptTokens), int32(usage.CompletionTokens), nil
}

// getThinking 获取思考内容。
//
// param: delta 变化数据。
// return: 思考内容。
func getThinking(delta *openai.ChatCompletionChunkChoiceDelta) *string {
	for _, name := range []string{"reasoning_content", "reasoning"} {
		var content string
		if err := json.Unmarshal([]byte(delta.JSON.ExtraFields[name].Raw()), &content); err == nil && content != "" {
			return &content
		}
	}
	return nil
}
