package server

import (
	"encoding/json/v2"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AnswerWriter 流式响应写入器。
type AnswerWriter struct {
	ctx *gin.Context
}

// NewAnswerWriter 创建流式响应写入器。
//
// param: ctx 上下文。
// return: 流式响应写入器。
func NewAnswerWriter(ctx *gin.Context) *AnswerWriter {
	// 切换为流式响应。
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.Status(http.StatusOK)
	ctx.Writer.Flush()
	// 返回流式响应写入器。
	return &AnswerWriter{ctx: ctx}
}

// Process 写入一个片段。
//
// param: kind 片段类型。
// param: content 片段内容。
// return: 错误信息。
func (writer *AnswerWriter) Process(kind string, content any) error {
	data, err := json.Marshal(map[string]any{"type": kind, "content": content})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer.ctx.Writer, "data: %s\n\n", data); err != nil {
		return err
	}
	writer.ctx.Writer.Flush()
	return nil
}

// Success 写入成功消息。
//
// param: session 会话编码。
// return: 错误信息。
func (writer *AnswerWriter) Success(session string) error {
	return writer.Process("finish", session)
}

// Failure 写入失败消息。
//
// param: message 异常消息。
// return: 错误信息。
func (writer *AnswerWriter) Failure(message string) error {
	return writer.Process("error", message)
}
