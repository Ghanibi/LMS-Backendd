package models

import "time"

type Subject struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:100;not null"`
	Code        string `gorm:"size:20;uniqueIndex;not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}