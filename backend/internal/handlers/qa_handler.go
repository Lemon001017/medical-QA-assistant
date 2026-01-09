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

// QAHandler handles question answering requests.
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

// AskStream handles streaming question answering requests via SSE.
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

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering
	c.Header("Access-Control-Allow-Origin", "*")

	// Use request context to handle client disconnection
	ctx := c.Request.Context()

	// Stream the response
	err := h.qaService.AskStream(ctx, userID.(uint), req.Question, func(chunk string) error {
		// Check if context is cancelled (client disconnected)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Format as SSE data
		data, err := json.Marshal(map[string]string{"chunk": chunk})
		if err != nil {
			return fmt.Errorf("failed to marshal chunk: %w", err)
		}

		// Write SSE format: data: {...}\n\n
		_, err = c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(data)))
		if err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}

		// Flush the response
		c.Writer.Flush()
		return nil
	})

	if err != nil {
		// Check if error is due to context cancellation (client disconnected)
		if err == context.Canceled || err == context.DeadlineExceeded {
			logger.L.Info("QA stream cancelled by client",
				zap.Uint("user_id", userID.(uint)),
			)
			return
		}

		// Send error as SSE event
		errorData, _ := json.Marshal(map[string]string{"error": err.Error()})
		c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(errorData)))
		c.Writer.Flush()
		return
	}

	// Send done event
	doneData, _ := json.Marshal(map[string]string{"done": "true"})
	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", string(doneData)))
	c.Writer.Flush()
}
