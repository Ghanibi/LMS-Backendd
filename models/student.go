package models

import "time"

type Student struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"` // Relasi ke tabel users
	NIS       string    `gorm:"size:30;uniqueIndex;not null" json:"nis"`
	NISN      string    `gorm:"size:30" json:"nisn"`
	Gender    string    `gorm:"size:20" json:"gender"` // <-- Tag json:"gender" memastikan data terkirim ke frontend
	ClassID   uint      `gorm:"not null" json:"class_id"`
	Class     Class     `gorm:"foreignKey:ClassID" json:"class"` // Relasi ke tabel classes
	Phone     string    `gorm:"size:20" json:"phone"`
	Address   string    `gorm:"size:255" json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}