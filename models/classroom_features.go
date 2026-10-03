package models

import "time"

type ClassAttendance struct {
	ID        uint      `gorm:"primaryKey"`
	ClassID   uint      `gorm:"not null;uniqueIndex:idx_class_attendance_daily,priority:1"`
	StudentID uint      `gorm:"not null;uniqueIndex:idx_class_attendance_daily,priority:2"`
	Date      time.Time `gorm:"type:date;not null;uniqueIndex:idx_class_attendance_daily,priority:3"`
	Status    string    `gorm:"size:20;not null"`
	Note      string    `gorm:"size:255"`
	TakenBy   uint      `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ClassroomEvent struct {
	ID          uint      `gorm:"primaryKey"`
	ClassID     uint      `gorm:"not null;index"`
	Title       string    `gorm:"size:150;not null"`
	Description string    `gorm:"type:text"`
	StartsAt    time.Time `gorm:"not null;index"`
	EndsAt      *time.Time
	CreatedBy   uint `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type HomeroomNote struct {
	ID        uint      `gorm:"primaryKey"`
	ClassID   uint      `gorm:"not null;index"`
	StudentID uint      `gorm:"not null;index"`
	AuthorID  uint      `gorm:"not null;index"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
}
