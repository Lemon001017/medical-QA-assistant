package handlers

import (
	"io"
	"medical-qa-assistant/internal/logger"
	"medical-qa-assistant/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DocumentHandler handles document related HTTP requests.
type DocumentHandler struct {
	documentService *services.DocumentService
}

func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

func (h *DocumentHandler) Create(c *gin.Context) {
	var req services.CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L.Warn("invalid create document request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		logger.L.Error("user id missing in context for document create")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	doc, err := h.documentService.Create(userID.(uint), &req)
	if err != nil {
		logger.L.Error("failed to create document",
			zap.Error(err),
			zap.Uint("user_id", userID.(uint)),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		logger.L.Error("user id missing in context for document list")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	docs, err := h.documentService.List(userID.(uint))
	if err != nil {
		logger.L.Error("failed to list documents",
			zap.Error(err),
			zap.Uint("user_id", userID.(uint)),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, docs)
}

func (h *DocumentHandler) Get(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		logger.L.Error("user id missing in context for document get")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idParam := c.Param("id")
	docID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logger.L.Warn("invalid document id in get",
			zap.String("id", idParam),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	doc, err := h.documentService.Get(userID.(uint), uint(docID))
	if err != nil {
		logger.L.Warn("document not found",
			zap.Error(err),
			zap.Uint("user_id", userID.(uint)),
			zap.Uint("document_id", uint(docID)),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		logger.L.Error("user id missing in context for document delete")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idParam := c.Param("id")
	docID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		logger.L.Warn("invalid document id in delete",
			zap.String("id", idParam),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	if err := h.documentService.Delete(userID.(uint), uint(docID)); err != nil {
		logger.L.Error("failed to delete document",
			zap.Error(err),
			zap.Uint("user_id", userID.(uint)),
			zap.Uint("document_id", uint(docID)),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Upload handles multipart document upload.
func (h *DocumentHandler) Upload(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		logger.L.Error("user id missing in context for document upload")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.L.Warn("file missing in upload request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	title := c.PostForm("title")
	if strings.TrimSpace(title) == "" {
		title = fileHeader.Filename
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.L.Error("failed to open uploaded file",
			zap.Error(err),
			zap.String("filename", fileHeader.Filename),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	contentBytes, err := io.ReadAll(file)
	if err != nil {
		logger.L.Error("failed to read uploaded file",
			zap.Error(err),
			zap.String("filename", fileHeader.Filename),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}

	req := services.CreateDocumentRequest{
		Title:   title,
		Content: string(contentBytes),
	}

	doc, err := h.documentService.Create(userID.(uint), &req)
	if err != nil {
		logger.L.Error("failed to create document from upload",
			zap.Error(err),
			zap.Uint("user_id", userID.(uint)),
			zap.String("title", title),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}
