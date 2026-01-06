package handlers

import (
	"context"
	"medical-qa-assistant/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var req services.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.qaService.Ask(context.Background(), userID.(uint), req.Question)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
