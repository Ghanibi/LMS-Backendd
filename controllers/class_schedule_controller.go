package controllers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"school-management/config"
	"school-management/models"
)

type classScheduleInput struct {
	Day               string `json:"day" binding:"required"`
	Period            int    `json:"period" binding:"required,min=1,max=12"`
	StartTime         string `json:"start_time" binding:"required"`
	EndTime           string `json:"end_time" binding:"required"`
	SubjectID         uint   `json:"subject_id"`
	TeacherID         *uint  `json:"teacher_id"`
	TeacherCode       int    `json:"teacher_code"`
	TeacherName       string `json:"teacher_name"`
	TeacherField      string `json:"teacher_field"`
	SubjectSuggestion string `json:"subject_suggestion"`
}

var scheduleDays = map[string]bool{"MONDAY": true, "TUESDAY": true, "WEDNESDAY": true, "THURSDAY": true, "FRIDAY": true, "SENIN": true, "SELASA": true, "RABU": true, "KAMIS": true, "JUMAT": true}

func GetClassSchedule(c *gin.Context) {
	classID := c.Param("id")
	var class models.Class
	if err := config.DB.First(&class, classID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kelas tidak ditemukan"})
		return
	}
	var schedule []models.ClassSchedule
	if err := config.DB.Preload("Subject").Preload("Teacher").Preload("Teacher.User").Where("class_id = ?", class.ID).Find(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil jadwal kelas"})
		return
	}
	dayOrder := map[string]int{"MONDAY": 1, "SENIN": 1, "TUESDAY": 2, "SELASA": 2, "WEDNESDAY": 3, "RABU": 3, "THURSDAY": 4, "KAMIS": 4, "FRIDAY": 5, "JUMAT": 5}
	sort.Slice(schedule, func(i, j int) bool {
		if dayOrder[schedule[i].Day] == dayOrder[schedule[j].Day] {
			return schedule[i].Period < schedule[j].Period
		}
		return dayOrder[schedule[i].Day] < dayOrder[schedule[j].Day]
	})
	c.JSON(http.StatusOK, gin.H{"data": schedule})
}

func ReplaceClassSchedule(c *gin.Context) {
	classID := c.Param("id")
	var class models.Class
	if err := config.DB.First(&class, classID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kelas tidak ditemukan"})
		return
	}
	var inputs []classScheduleInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format jadwal tidak valid: " + err.Error()})
		return
	}
	seen := make(map[string]bool)
	entries := make([]models.ClassSchedule, 0, len(inputs))
	for _, input := range inputs {
		day := strings.ToUpper(strings.TrimSpace(input.Day))
		if !scheduleDays[day] || input.Period < 1 || input.Period > 12 || input.StartTime >= input.EndTime {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Hari, jam, mata pelajaran, atau nomor jam tidak valid"})
			return
		}
		key := day + ":" + strconv.Itoa(input.Period)
		if seen[key] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ada dua mata pelajaran pada hari dan jam yang sama"})
			return
		}
		seen[key] = true
		var subjectID *uint
		if input.SubjectID > 0 {
			var subject models.Subject
			if err := config.DB.First(&subject, input.SubjectID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Mata pelajaran tidak ditemukan"})
				return
			}
			subjectID = &input.SubjectID
		}
		if input.TeacherID != nil && *input.TeacherID > 0 {
			var teacher models.Teacher
			if err := config.DB.First(&teacher, *input.TeacherID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Guru tidak ditemukan"})
				return
			}
		} else {
			input.TeacherID = nil
		}
		entries = append(entries, models.ClassSchedule{ClassID: class.ID, Day: day, Period: input.Period, StartTime: input.StartTime, EndTime: input.EndTime, SubjectID: subjectID, TeacherID: input.TeacherID, TeacherCode: input.TeacherCode, TeacherName: strings.TrimSpace(input.TeacherName), TeacherField: strings.TrimSpace(input.TeacherField), SubjectSuggestion: strings.TrimSpace(input.SubjectSuggestion), NeedsReview: subjectID == nil})
	}
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("class_id = ?", class.ID).Delete(&models.ClassSchedule{}).Error; err != nil {
			return err
		}
		if len(entries) > 0 {
			return tx.Create(&entries).Error
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Jadwal gagal disimpan"})
		return
	}
	GetClassSchedule(c)
}
