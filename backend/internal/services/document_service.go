package services

import (
	"context"
	"errors"

	"medical-qa-assistant/internal/logger"
	"medical-qa-assistant/internal/models"
	"medical-qa-assistant/internal/repositories"

	"go.uber.org/zap"
)

// DocumentService 包含文档管理的业务逻辑
type DocumentService struct {
	documentRepo *repositories.DocumentRepository
	ragService   *RAGService
}

func NewDocumentService(documentRepo *repositories.DocumentRepository, ragService *RAGService) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
		ragService:   ragService,
	}
}

type CreateDocumentRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=255"`
	Content string `json:"content" binding:"required"`
}

type DocumentResponse struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *DocumentService) Create(userID uint, req *CreateDocumentRequest) (*models.Document, error) {
	if userID == 0 {
		return nil, errors.New("invalid user")
	}

	doc := &models.Document{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Status:  "ready",
	}

	// 先保存文档到数据库，获得 ID
	if err := s.documentRepo.Create(doc); err != nil {
		logger.L.Error("failed to create document",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.String("title", req.Title),
		)
		return nil, err
	}

	if s.ragService != nil && s.ragService.IsEnabled() {
		if err := s.ragService.IndexDocument(context.Background(), doc); err != nil {
			logger.L.Error("failed to index document into RAG",
				zap.Error(err),
				zap.Uint("document_id", doc.ID),
				zap.Uint("user_id", doc.UserID),
			)
			doc.Status = "indexing_failed"
			s.documentRepo.Update(doc)
		} else {
			logger.L.Info("document indexed into RAG",
				zap.Uint("document_id", doc.ID),
				zap.Uint("user_id", doc.UserID),
			)
		}
	}

	return doc, nil
}

func (s *DocumentService) List(userID uint) ([]models.Document, error) {
	if userID == 0 {
		return nil, errors.New("invalid user")
	}
	return s.documentRepo.ListByUser(userID)
}

func (s *DocumentService) Get(userID, docID uint) (*models.Document, error) {
	if userID == 0 {
		return nil, errors.New("invalid user")
	}
	return s.documentRepo.GetByIDAndUser(docID, userID)
}

func (s *DocumentService) Delete(userID, docID uint) error {
	if userID == 0 {
		return errors.New("invalid user")
	}

	// 先删除 Chroma 中的向量数据（如果启用）
	if s.ragService != nil && s.ragService.IsEnabled() {
		if err := s.ragService.DeleteDocument(context.Background(), docID, userID); err != nil {
			logger.L.Error("failed to delete document from RAG",
				zap.Error(err),
				zap.Uint("document_id", docID),
				zap.Uint("user_id", userID),
			)
		}
	}

	// 删除数据库中的文档记录
	if err := s.documentRepo.DeleteByIDAndUser(docID, userID); err != nil {
		logger.L.Error("failed to delete document from database",
			zap.Error(err),
			zap.Uint("document_id", docID),
			zap.Uint("user_id", userID),
		)
		return err
	}

	logger.L.Info("document deleted successfully",
		zap.Uint("document_id", docID),
		zap.Uint("user_id", userID),
	)

	return nil
}
