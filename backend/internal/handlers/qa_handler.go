package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"medical-qa-assistant/internal/logger"
	"medical-qa-assistant/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// QAHandler 处理问答请求
type QAHandler struct {
	qaService *services.QAService
}

func NewQAHandler(qaService *services.QAService) *QAHandler {
	return &QAHandler{qaService: qaService}
}

func (h *QAHandler) Ask(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		logger.L.Error("user id missing in context for QA Ask")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var req services.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Warn("invalid QA Ask request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.qaService.Ask(context.Background(), userID.(uint), req.Question)
	if err != nil {
		logger.L.Error("QA Ask failed",
			zap.Error(err),
			zap.Uint("user_id", userID.(uint)),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AskStream 通过 SSE 处理流式问答请求
func (h *QAHandler) AskStream(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		logger.L.Error("user id missing in context for QA AskStream")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var req services.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Warn("invalid QA AskStream request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用 nginx 缓冲
	c.Header("Access-Control-Allow-Origin", "*")

	// 使用请求上下文处理客户端断开连接
	ctx := c.Request.Context()

	// 流式传输响应
	err := h.qaService.AskStream(ctx, userID.(uint), req.Question, func(chunk string) error {
		// 检查上下文是否已取消（客户端断开连接）
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 格式化为 SSE 数据
		data, err := json.Marshal(map[string]string{"chunk": chunk})
		if err != nil {
			return fmt.Errorf("failed to marshal chunk: %w", err)
		}

		// 写入 SSE 格式：data: {...}\n\n
		_, err = c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(data)))
		if err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}

		// 刷新响应
		c.Writer.Flush()
		return nil
	})

	if err != nil {
		// 检查错误是否由于上下文取消（客户端断开连接）
		if err == context.Canceled || err == context.DeadlineExceeded {
			logger.L.Info("QA stream cancelled by client",
				zap.Uint("user_id", userID.(uint)),
			)
			return
		}

		// 将错误作为 SSE 事件发送
		errorData, _ := json.Marshal(map[string]string{"error": err.Error()})
		c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(errorData)))
		c.Writer.Flush()
		return
	}

	// 发送完成事件
	doneData, _ := json.Marshal(map[string]string{"done": "true"})
	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(doneData)))
	c.Writer.Flush()
}
