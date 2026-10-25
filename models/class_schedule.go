package models

import "time"

// ClassSchedule stores one subject in one class timetable slot.
type ClassSchedule struct {
	ID                uint      `gorm:"primaryKey"`
	ClassID           uint      `gorm:"not null;uniqueIndex:idx_class_schedule_slot" json:"class_id"`
	Class             Class     `gorm:"foreignKey:ClassID" json:"-"`
	Day               string    `gorm:"size:12;not null;uniqueIndex:idx_class_schedule_slot" json:"day"`
	Period            int       `gorm:"not null;uniqueIndex:idx_class_schedule_slot" json:"period"`
	StartTime         string    `gorm:"size:5;not null" json:"start_time"`
	EndTime           string    `gorm:"size:5;not null" json:"end_time"`
	SubjectID         *uint     `json:"subject_id,omitempty"`
	Subject           *Subject  `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	TeacherID         *uint     `json:"teacher_id,omitempty"`
	Teacher           *Teacher  `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	TeacherCode       int       `json:"teacher_code"`
	TeacherName       string    `gorm:"size:150" json:"teacher_name"`
	TeacherField      string    `gorm:"size:255" json:"teacher_field"`
	SubjectSuggestion string    `gorm:"type:text" json:"subject_suggestion"`
	NeedsReview       bool      `gorm:"not null;default:false" json:"needs_review"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
