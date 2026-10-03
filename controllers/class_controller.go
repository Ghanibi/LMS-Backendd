package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"school-management/config"
	"school-management/models"
)

type ClassInput struct {
	Name              string `json:"name" binding:"required"`
	EducationLevelID  uint   `json:"education_level_id" binding:"required"`
	HomeroomTeacherID *uint  `json:"homeroom_teacher_id"`
	Grade             int    `json:"grade" binding:"required"`
	Major             string `json:"major"`
	ClassNumber       *int   `json:"class_number"`
	IsPlus            bool   `json:"is_plus"`
}

// validateClassInput memeriksa jenjang, rentang tingkat kelas, dan jurusan.
// Mengembalikan pesan error (string kosong jika valid).
//   - SMP: tingkat 7-9
//   - SMA: tingkat 10-12
//   - SMK: tingkat 10-12 dan jurusan wajib diisi
func validateClassInput(input ClassInput) string {
	var level models.EducationLevel
	if err := config.DB.First(&level, input.EducationLevelID).Error; err != nil {
		return "Jenjang pendidikan tidak ditemukan"
	}

	levelName := strings.ToUpper(level.Name)
	switch levelName {
	case "SMP":
		if input.Grade < 7 || input.Grade > 9 {
			return "Tingkat kelas SMP harus 7 sampai 9"
		}
	case "SMA", "SMK":
		if input.Grade < 10 || input.Grade > 12 {
			return "Tingkat kelas " + levelName + " harus 10 sampai 12"
		}
	}

	if levelName == "SMK" && strings.TrimSpace(input.Major) == "" {
		return "Jurusan wajib diisi untuk SMK"
	}

	return ""
}

func homeroomTeacherConflict(teacherID *uint, classID uint) bool {
	if teacherID == nil || *teacherID == 0 {
		return false
	}
	query := config.DB.Model(&models.Class{}).Where("homeroom_teacher_id = ?", *teacherID)
	if classID != 0 {
		query = query.Where("id <> ?", classID)
	}
	var count int64
	query.Count(&count)
	return count > 0
}

func GetClasses(c *gin.Context) {
	var classes []models.Class
	if err := config.DB.Preload("EducationLevel").Preload("HomeroomTeacher").Preload("HomeroomTeacher.User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).Find(&classes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kelas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data kelas", "data": classes})
}

func GetClassByID(c *gin.Context) {
	id := c.Param("id")
	var class models.Class
	if err := config.DB.Preload("EducationLevel").First(&class, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Kelas tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil detail kelas", "data": class})
}

func CreateClass(c *gin.Context) {
	var input ClassInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg := validateClassInput(input); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if input.HomeroomTeacherID != nil {
		var teacher models.Teacher
		if err := config.DB.First(&teacher, *input.HomeroomTeacherID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Guru wali kelas tidak ditemukan"})
			return
		}
		if homeroomTeacherConflict(input.HomeroomTeacherID, 0) {
			c.JSON(http.StatusConflict, gin.H{"error": "Guru tersebut sudah menjadi wali kelas di kelas lain"})
			return
		}
	}

	class := models.Class{
		Name:              input.Name,
		EducationLevelID:  input.EducationLevelID,
		HomeroomTeacherID: input.HomeroomTeacherID,
		Grade:             input.Grade,
		Major:             input.Major,
		ClassNumber:       input.ClassNumber,
		IsPlus:            input.IsPlus,
	}

	if err := config.DB.Create(&class).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat kelas, pastikan nama kelas unik"})
		return
	}

	config.DB.Preload("EducationLevel").First(&class, class.ID)
	c.JSON(http.StatusCreated, gin.H{"message": "Kelas berhasil ditambahkan", "data": class})
}

func UpdateClass(c *gin.Context) {
	id := c.Param("id")
	var class models.Class

	if err := config.DB.First(&class, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kelas tidak ditemukan"})
		return
	}

	var input ClassInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if msg := validateClassInput(input); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if input.HomeroomTeacherID != nil {
		var teacher models.Teacher
		if err := config.DB.First(&teacher, *input.HomeroomTeacherID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Guru wali kelas tidak ditemukan"})
			return
		}
		if homeroomTeacherConflict(input.HomeroomTeacherID, class.ID) {
			c.JSON(http.StatusConflict, gin.H{"error": "Guru tersebut sudah menjadi wali kelas di kelas lain"})
			return
		}
	}

	class.Name = input.Name
	class.EducationLevelID = input.EducationLevelID
	class.HomeroomTeacherID = input.HomeroomTeacherID
	class.Grade = input.Grade
	class.Major = input.Major
	class.ClassNumber = input.ClassNumber
	class.IsPlus = input.IsPlus

	if err := config.DB.Save(&class).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui kelas, pastikan nama kelas unik"})
		return
	}

	config.DB.Preload("EducationLevel").Preload("HomeroomTeacher").Preload("HomeroomTeacher.User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name")
	}).First(&class, class.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Kelas berhasil diperbarui", "data": class})
}

func DeleteClass(c *gin.Context) {
	id := c.Param("id")
	var class models.Class
	if err := config.DB.First(&class, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kelas tidak ditemukan"})
		return
	}
	if err := config.DB.Delete(&class).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kelas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Kelas berhasil dihapus"})
}
