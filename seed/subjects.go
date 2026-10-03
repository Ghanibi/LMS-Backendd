package seed

import (
	"fmt"
	"log"

	"school-management/config"
	"school-management/models"
)

var defaultSubjects = []string{
	"Matematika",
	"Bahasa Indonesia",
	"Bahasa Inggris",
	"Pendidikan Agama",
	"PPKn",
	"Sejarah Indonesia",
	"Pendidikan Jasmani, Olahraga, dan Kesehatan",
	"Seni Budaya",
	"IPAS",
	"Projek Kreatif dan Kewirausahaan",
	"Bimbingan Konseling",
	"Pemrograman Web",
	"Basis Data",
	"Pemrograman Berorientasi Objek",
	"Desain Grafis",
	"Jaringan Komputer",
	"Administrasi Sistem Jaringan",
	"Marketing Digital",
	"Manajemen Perkantoran",
}

func SeedSubjects() {
	for idx, name := range defaultSubjects {
		var existing models.Subject
		if err := config.DB.Where("name = ?", name).First(&existing).Error; err == nil {
			continue // sudah ada, lewati
		}

		subject := models.Subject{
			Name:        name,
			Code:        fmt.Sprintf("MP-%d", 101+idx),
			Description: "Mata pelajaran wajib dan kejuruan sekolah",
		}
		if err := config.DB.Create(&subject).Error; err != nil {
			log.Printf("Gagal seed mata pelajaran %s: %v", name, err)
		}
	}

	log.Println("Seeder Mata Pelajaran selesai.")
}