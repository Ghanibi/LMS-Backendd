package models

import "time"

type TeacherSubject struct {
	ID        uint    `gorm:"primaryKey"`
	TeacherID uint    `gorm:"not null;uniqueIndex:idx_teacher_subject_class"`
	Teacher   Teacher `gorm:"foreignKey:TeacherID"`
	SubjectID uint    `gorm:"not null;uniqueIndex:idx_teacher_subject_class"`
	Subject   Subject `gorm:"foreignKey:SubjectID"`
	ClassID   uint    `gorm:"not null;uniqueIndex:idx_teacher_subject_class"`
	Class     Class   `gorm:"foreignKey:ClassID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}