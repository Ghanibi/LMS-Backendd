package seed

import (
	"fmt"

	"school-management/config"
	"school-management/models"
)

func SeedEducationLevels() {
	levels := []models.EducationLevel{
		{Name: "SMP"},
		{Name: "SMA"},
		{Name: "SMK"},
	}

	for _, level := range levels {
		var existing models.EducationLevel

		result := config.DB.
			Where("name = ?", level.Name).
			First(&existing)

		if result.Error != nil {
			config.DB.Create(&level)
			fmt.Println("Education level dibuat:", level.Name)
		}
	}
}