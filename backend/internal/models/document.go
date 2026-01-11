package models

import (
	"time"
)

// Document 存储用户上传的医学文档内容和元数据
type Document struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null"`
	Content   string    `json:"content" gorm:"type:longtext;not null"`
	Status    string    `json:"status" gorm:"type:varchar(50);default:ready"` // ready, processing, failed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
