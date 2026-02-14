// Package models for User data
package models

import "time"

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"column:email;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;not null"`
	IsActive     bool      `json:"is_active" gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}
