package models

import "time"

type Announcement struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:150;not null"`
	Content     string `gorm:"type:text;not null"`
	TargetRole  string `gorm:"size:20;not null;default:'Semua'"`
	ClassID     *uint  `gorm:"index" json:"class_id,omitempty"`
	AuthorID    uint   `gorm:"not null"`
	IsPublished bool   `gorm:"default:false"`
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
