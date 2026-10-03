package models

import "time"

type Class struct {
	ID                uint           `gorm:"primaryKey"`
	Name              string         `gorm:"size:100;uniqueIndex;not null"`
	EducationLevelID  uint           `gorm:"not null"`
	EducationLevel    EducationLevel `gorm:"foreignKey:EducationLevelID"` // BARU: Relasi ke tabel education_levels
	HomeroomTeacherID *uint          `gorm:"uniqueIndex:idx_classes_single_homeroom_teacher" json:"homeroom_teacher_id,omitempty"`
	HomeroomTeacher   *Teacher       `gorm:"foreignKey:HomeroomTeacherID" json:"homeroom_teacher,omitempty"`
	Grade             int            `gorm:"not null"`
	Major             string         `gorm:"size:50"`
	ClassNumber       *int
	IsPlus            bool `gorm:"default:false"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
