package models

import "time"

type Teacher struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"` // Relasi ke tabel users
	NIP       string    `gorm:"size:30;uniqueIndex" json:"nip"`
	Gender    string    `gorm:"size:20" json:"gender"`
	Subject   string    `gorm:"size:100" json:"subject"` // Bidang Studi / Mata Pelajaran
	Phone     string    `gorm:"size:20" json:"phone"`
	Address   string    `gorm:"size:255" json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}