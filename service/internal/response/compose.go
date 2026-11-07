package response

import (
	"service/internal/service"

	"github.com/gin-gonic/gin"
)

// ComposeAnswer 编排回答问题。
func ComposeAnswer(ctx *gin.Context, session *string, content *string, novels []string, callback func(kind string, content string) error) (*string, error) {
	// 保存用户提问。
	sessionId, _, _, err := service.Converse.AppendHumanMessage(ctx, session, content, novels)
	if err != nil {
		return nil, err
	}

	return sessionId, nil
}
