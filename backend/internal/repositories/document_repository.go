package repositories

import (
	"medical-qa-assistant/internal/models"

	"gorm.io/gorm"
)

// DocumentRepository provides CRUD operations for documents.
type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(doc *models.Document) error {
	return r.db.Create(doc).Error
}

func (r *DocumentRepository) ListByUser(userID uint) ([]models.Document, error) {
	var docs []models.Document
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *DocumentRepository) GetByIDAndUser(id, userID uint) (*models.Document, error) {
	var doc models.Document
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *DocumentRepository) DeleteByIDAndUser(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Document{}).Error
}
