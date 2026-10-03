package seed

import (
	"log"
	"time"

	"school-management/config"
	"school-management/models"
)

func SeedCalendars() {
	var count int64
	config.DB.Model(&models.AcademicEvent{}).Count(&count)
	if count > 0 {
		return
	}

	uts := time.Date(2026, 10, 10, 0, 0, 0, 0, time.Local)
	libur := time.Date(2026, 11, 10, 0, 0, 0, 0, time.Local)

	events := []models.AcademicEvent{
		{
			Title:       "Ujian Tengah Semester (UTS)",
			Description: "Pelaksanaan UTS Semester Ganjil",
			StartDate:   uts,
			EndDate:     uts,
			Type:        "Akademik",
		},
		{
			Title:       "Libur Nasional Hari Pahlawan",
			Description: "Libur kegiatan belajar mengajar",
			StartDate:   libur,
			EndDate:     libur,
			Type:        "Libur",
		},
	}

	for _, ev := range events {
		if err := config.DB.Create(&ev).Error; err != nil {
			log.Printf("Gagal seed agenda %s: %v", ev.Title, err)
		}
	}

	log.Println("Seeder Kalender selesai.")
}