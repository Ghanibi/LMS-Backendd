package models

import "time"

type ClassSubject struct {
	ID        uint    `gorm:"primaryKey"`
	ClassID   uint    `gorm:"not null;uniqueIndex:idx_class_subject"`
	Class     Class   `gorm:"foreignKey:ClassID"`
	SubjectID uint    `gorm:"not null;uniqueIndex:idx_class_subject"`
	Subject   Subject `gorm:"foreignKey:SubjectID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}