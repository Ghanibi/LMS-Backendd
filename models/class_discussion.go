package models

import "time"

type ClassDiscussion struct {
	ID        uint   `gorm:"primaryKey"`
	ClassID   uint   `gorm:"not null;index"`
	AuthorID  uint   `gorm:"not null;index"`
	ParentID  *uint  `gorm:"index"`
	Title     string `gorm:"size:150"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Author    User `gorm:"foreignKey:AuthorID"`
}
