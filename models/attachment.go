package models

import "time"

type Attachment struct {
	ID           uint   `gorm:"primaryKey"`
	ResourceType string `gorm:"size:40;not null;index:idx_resource_attachment"`
	ResourceID   uint   `gorm:"not null;index:idx_resource_attachment"`
	FileName     string `gorm:"size:255;not null"`
	FileURL      string `gorm:"size:500;not null"`
	UploadedBy   uint   `gorm:"not null"`
	CreatedAt    time.Time
}
