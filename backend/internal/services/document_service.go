package services

import (
	"errors"
	"medical-qa-assistant/internal/models"
	"medical-qa-assistant/internal/repositories"
)

// DocumentService contains business logic for document management.
type DocumentService struct {
	documentRepo *repositories.DocumentRepository
}

func NewDocumentService(documentRepo *repositories.DocumentRepository) *DocumentService {
	return &DocumentService{documentRepo: documentRepo}
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

	if err := s.documentRepo.Create(doc); err != nil {
		return nil, err
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
	return s.documentRepo.DeleteByIDAndUser(docID, userID)
}
