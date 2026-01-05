package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"type:varchar(191);uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"type:varchar(191);uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"type:varchar(255);not null"`
	Role      string         `json:"role" gorm:"type:varchar(50);default:user"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)
