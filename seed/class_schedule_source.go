package seed

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"strings"

	"school-management/config"
	"school-management/models"
)

//go:embed data/schedule_source_2026_2027.json
var scheduleSourceJSON []byte

type sourceScheduleFile struct {
	SchoolYear string                `json:"school_year"`
	Classes    []string              `json:"classes"`
	Entries    []sourceScheduleEntry `json:"entries"`
}

type sourceScheduleEntry struct {
	ClassName         string `json:"class_name"`
	Day               string `json:"day"`
	Period            int    `json:"period"`
	StartTime         string `json:"start_time"`
	EndTime           string `json:"end_time"`
	TeacherCode       int    `json:"teacher_code"`
	TeacherName       string `json:"teacher_name"`
	TeacherField      string `json:"teacher_field"`
	SubjectName       string `json:"subject_name"`
	SubjectSuggestion string `json:"subject_suggestion"`
	NeedsReview       bool   `json:"needs_review"`
}

func scheduleClassKey(name string) string {
	name = strings.ToUpper(strings.Join(strings.Fields(name), " "))
	name = strings.ReplaceAll(name, " BDR ", " PEMASARAN ")
	return name
}

func scheduleSubjectCode(name string) string {
	return fmt.Sprintf("SRC-%08X", crc32.ChecksumIEEE([]byte(strings.ToLower(strings.TrimSpace(name)))))
}

// SeedClassScheduleSource imports the official 2026/2027 timetable once. Future
// server starts preserve any timetable already saved or edited by Admin.
func SeedClassScheduleSource() {
	var existing int64
	if err := config.DB.Model(&models.ClassSchedule{}).Count(&existing).Error; err != nil {
		log.Printf("Gagal memeriksa jadwal kelas: %v", err)
		return
	}
	if existing > 0 {
		return
	}
	var source sourceScheduleFile
	if err := json.Unmarshal(scheduleSourceJSON, &source); err != nil {
		log.Printf("Gagal membaca jadwal sumber: %v", err)
		return
	}
	var smk models.EducationLevel
	if err := config.DB.Where("name = ?", "SMK").FirstOrCreate(&smk, models.EducationLevel{Name: "SMK"}).Error; err != nil {
		log.Printf("Gagal menyiapkan jenjang SMK: %v", err)
		return
	}

	classes := make(map[string]models.Class, len(source.Classes))
	var allClasses []models.Class
	if err := config.DB.Find(&allClasses).Error; err != nil {
		log.Printf("Gagal membaca kelas: %v", err)
		return
	}
	for _, sourceName := range source.Classes {
		key := scheduleClassKey(sourceName)
		var class models.Class
		for _, candidate := range allClasses {
			if scheduleClassKey(candidate.Name) == key {
				class = candidate
				break
			}
		}
		if class.ID == 0 {
			parts := strings.Fields(key)
			if len(parts) < 2 {
				continue
			}
			grade := 0
			fmt.Sscanf(parts[0], "%d", &grade)
			major := parts[1]
			classNumber := 0
			isPlus := strings.Contains(key, " PLUS")
			if !isPlus {
				fmt.Sscanf(parts[len(parts)-1], "%d", &classNumber)
			}
			classNumberPtr := &classNumber
			if isPlus {
				classNumberPtr = nil
			}
			class = models.Class{Name: strings.Join(parts, " "), EducationLevelID: smk.ID, Grade: grade, Major: major, IsPlus: isPlus, ClassNumber: classNumberPtr}
			if err := config.DB.Where("name = ?", class.Name).FirstOrCreate(&class, class).Error; err != nil {
				log.Printf("Gagal menyiapkan kelas %s: %v", class.Name, err)
				return
			}
		}
		classes[sourceName] = class
	}

	entries := make([]models.ClassSchedule, 0, len(source.Entries))
	subjectCache := make(map[string]models.Subject)
	for _, item := range source.Entries {
		class, ok := classes[item.ClassName]
		if !ok {
			continue
		}
		entry := models.ClassSchedule{
			ClassID: class.ID, Day: item.Day, Period: item.Period,
			StartTime: item.StartTime, EndTime: item.EndTime,
			TeacherCode: item.TeacherCode, TeacherName: item.TeacherName,
			TeacherField: item.TeacherField, SubjectSuggestion: item.SubjectSuggestion,
			NeedsReview: item.NeedsReview,
		}
		if item.SubjectName != "" {
			cacheKey := strings.ToLower(strings.TrimSpace(item.SubjectName))
			subject, exists := subjectCache[cacheKey]
			if !exists {
				err := config.DB.Where("LOWER(name) = LOWER(?)", item.SubjectName).First(&subject).Error
				if err != nil {
					subject = models.Subject{Name: item.SubjectName, Code: scheduleSubjectCode(item.SubjectName), Description: "Diimpor dari jadwal SMK tahun pelajaran 2026/2027"}
					if err := config.DB.Where("code = ?", subject.Code).FirstOrCreate(&subject, subject).Error; err != nil {
						log.Printf("Gagal menyiapkan mata pelajaran %s: %v", item.SubjectName, err)
						continue
					}
				}
				subjectCache[cacheKey] = subject
			}
			entry.SubjectID = &subject.ID
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		log.Println("Jadwal sumber tidak berisi slot yang bisa diimpor.")
		return
	}
	if err := config.DB.CreateInBatches(&entries, 200).Error; err != nil {
		log.Printf("Gagal mengimpor jadwal kelas: %v", err)
		return
	}
	log.Printf("Jadwal %s diimpor: %d slot untuk %d kelas; slot tanpa mapel pasti ditandai untuk ditinjau.", source.SchoolYear, len(entries), len(classes))
}
