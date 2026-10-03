package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"school-management/config"
	"school-management/models"
)

type MaterialInput struct {
	ClassSubjectID uint   `json:"class_subject_id" binding:"required"`
	TeacherID      uint   `json:"teacher_id"`
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description"`
	FileURL        string `json:"file_url"`
}

// teacherOwnsClassSubject memastikan guru benar-benar ditugaskan mengajar
// kelas & mapel yang terkait dengan ClassSubjectID tersebut.
func teacherOwnsClassSubject(teacherID uint, classSubjectID uint) bool {
	var cs models.ClassSubject
	if err := config.DB.First(&cs, classSubjectID).Error; err != nil {
		return false
	}
	var count int64
	config.DB.Model(&models.TeacherSubject{}).
		Where("teacher_id = ? AND class_id = ? AND subject_id = ?", teacherID, cs.ClassID, cs.SubjectID).
		Count(&count)
	return count > 0
}

func GetMaterials(c *gin.Context) {
	query := config.DB.
		Preload("Teacher").Preload("Teacher.User").
		Preload("ClassSubject").Preload("ClassSubject.Class").Preload("ClassSubject.Subject")

	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Data guru tidak ditemukan"})
			return
		}
		query = query.Where("teacher_id = ?", teacher.ID)
	}

	var materials []models.Material
	if err := query.Order("created_at desc").Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data materi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data materi", "data": materials})
}

func GetMaterialByID(c *gin.Context) {
	id := c.Param("id")
	var material models.Material
	if err := config.DB.Preload("Teacher").Preload("ClassSubject").First(&material, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Materi tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}

	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil || material.TeacherID != teacher.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses ke materi ini"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil detail materi", "data": material})
}

func CreateMaterial(c *gin.Context) {
	var input MaterialInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacherID := input.TeacherID

	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Data guru tidak ditemukan"})
			return
		}
		teacherID = teacher.ID
		if !teacherOwnsClassSubject(teacher.ID, input.ClassSubjectID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak ditugaskan mengajar kelas & mapel ini"})
			return
		}
	} else if teacherID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guru wajib dipilih"})
		return
	}

	material := models.Material{
		ClassSubjectID: input.ClassSubjectID,
		TeacherID:      teacherID,
		Title:          input.Title,
		Description:    input.Description,
		FileURL:        input.FileURL,
	}

	if err := config.DB.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat materi"})
		return
	}

	config.DB.Preload("Teacher").Preload("ClassSubject").First(&material, material.ID)
	c.JSON(http.StatusCreated, gin.H{"message": "Materi berhasil ditambahkan", "data": material})
}

func UpdateMaterial(c *gin.Context) {
	id := c.Param("id")
	var material models.Material

	if err := config.DB.First(&material, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Materi tidak ditemukan"})
		return
	}

	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil || material.TeacherID != teacher.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses ke materi ini"})
			return
		}
	}

	var input MaterialInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if currentRole(c) == "TEACHER" {
		teacher, _ := currentTeacher(c)
		if !teacherOwnsClassSubject(teacher.ID, input.ClassSubjectID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak ditugaskan mengajar kelas & mapel ini"})
			return
		}
	} else if input.TeacherID != 0 {
		material.TeacherID = input.TeacherID
	}

	material.ClassSubjectID = input.ClassSubjectID
	material.Title = input.Title
	material.Description = input.Description
	material.FileURL = input.FileURL

	if err := config.DB.Save(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui materi"})
		return
	}

	config.DB.Preload("Teacher").Preload("ClassSubject").First(&material, material.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Materi berhasil diperbarui", "data": material})
}

func DeleteMaterial(c *gin.Context) {
	id := c.Param("id")
	var material models.Material
	if err := config.DB.First(&material, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Materi tidak ditemukan"})
		return
	}

	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil || material.TeacherID != teacher.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki akses ke materi ini"})
			return
		}
	}

	if err := config.DB.Delete(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus materi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Materi berhasil dihapus"})
}