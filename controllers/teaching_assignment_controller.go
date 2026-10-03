package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

type TeachingAssignmentInput struct {
	TeacherID uint `json:"teacher_id" binding:"required"`
	ClassID   uint `json:"class_id" binding:"required"`
	SubjectID uint `json:"subject_id" binding:"required"`
}

func GetTeachingAssignments(c *gin.Context) {
	var items []models.TeacherSubject
	if err := config.DB.Preload("Teacher").Preload("Teacher.User").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data penugasan mengajar"})
		return
	}

	type out struct {
		ID          uint   `json:"ID"`
		TeacherID   uint   `json:"TeacherID"`
		TeacherName string `json:"TeacherName"`
		ClassID     uint   `json:"ClassID"`
		ClassName   string `json:"ClassName"`
		SubjectID   uint   `json:"SubjectID"`
		SubjectName string `json:"SubjectName"`
	}

	result := make([]out, 0, len(items))
	for _, it := range items {
		var class models.Class
		var subject models.Subject
		config.DB.First(&class, it.ClassID)
		config.DB.First(&subject, it.SubjectID)

		result = append(result, out{
			ID:          it.ID,
			TeacherID:   it.TeacherID,
			TeacherName: it.Teacher.User.Name,
			ClassID:     it.ClassID,
			ClassName:   class.Name,
			SubjectID:   it.SubjectID,
			SubjectName: subject.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data penugasan mengajar", "data": result})
}

func CreateTeachingAssignment(c *gin.Context) {
	var input TeachingAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var teacher models.Teacher
	if err := config.DB.First(&teacher, input.TeacherID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guru tidak ditemukan"})
		return
	}
	var class models.Class
	if err := config.DB.First(&class, input.ClassID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kelas tidak ditemukan"})
		return
	}
	var subject models.Subject
	if err := config.DB.First(&subject, input.SubjectID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mata pelajaran tidak ditemukan"})
		return
	}

	var count int64
	config.DB.Model(&models.TeacherSubject{}).
		Where("teacher_id = ? AND class_id = ? AND subject_id = ?", input.TeacherID, input.ClassID, input.SubjectID).
		Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guru ini sudah ditugaskan untuk kelas & mapel tersebut"})
		return
	}

	// Pastikan pasangan Kelas+Mapel (ClassSubject) tersedia, buat jika belum ada
	var classSubject models.ClassSubject
	if err := config.DB.Where("class_id = ? AND subject_id = ?", input.ClassID, input.SubjectID).First(&classSubject).Error; err != nil {
		classSubject = models.ClassSubject{ClassID: input.ClassID, SubjectID: input.SubjectID}
		if err := config.DB.Create(&classSubject).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat pasangan kelas & mapel"})
			return
		}
	}

	teacherSubject := models.TeacherSubject{
		TeacherID: input.TeacherID,
		ClassID:   input.ClassID,
		SubjectID: input.SubjectID,
	}
	if err := config.DB.Create(&teacherSubject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat penugasan mengajar"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Penugasan mengajar berhasil ditambahkan"})
}

func DeleteTeachingAssignment(c *gin.Context) {
	id := c.Param("id")
	var item models.TeacherSubject
	if err := config.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Penugasan mengajar tidak ditemukan"})
		return
	}
	if err := config.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus penugasan mengajar"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Penugasan mengajar berhasil dihapus"})
}