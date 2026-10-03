package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"school-management/config"
	"school-management/models"
)

type AssignmentInput struct {
	ClassSubjectID uint      `json:"class_subject_id" binding:"required"`
	TeacherID      uint      `json:"teacher_id"`
	Title          string    `json:"title" binding:"required"`
	Description    string    `json:"description"`
	DueDate        time.Time `json:"due_date" binding:"required"`
	MaxScore       float64   `json:"max_score"`
}

func GetAssignments(c *gin.Context) {
	query := config.DB.Preload("Teacher")
	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Data guru tidak ditemukan"})
			return
		}
		query = query.Where("teacher_id = ?", teacher.ID)
	}
	var assignments []models.Assignment
	if err := query.Find(&assignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data tugas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data tugas", "data": assignments})
}

func GetAssignmentByID(c *gin.Context) {
	id := c.Param("id")
	var assignment models.Assignment
	if err := config.DB.Preload("Teacher").First(&assignment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tugas tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil detail tugas", "data": assignment})
}

func CreateAssignment(c *gin.Context) {
	var input AssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Data guru tidak ditemukan"})
			return
		}
		if !teacherOwnsClassSubject(teacher.ID, input.ClassSubjectID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak ditugaskan mengajar kelas dan mata pelajaran ini"})
			return
		}
		input.TeacherID = teacher.ID
	} else if input.TeacherID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guru wajib dipilih"})
		return
	}

	maxScore := input.MaxScore
	if maxScore == 0 {
		maxScore = 100 // Set default max score
	}

	assignment := models.Assignment{
		ClassSubjectID: input.ClassSubjectID,
		TeacherID:      input.TeacherID,
		Title:          input.Title,
		Description:    input.Description,
		DueDate:        input.DueDate,
		MaxScore:       maxScore,
	}

	if err := config.DB.Create(&assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat tugas"})
		return
	}

	config.DB.Preload("Teacher").First(&assignment, assignment.ID)
	c.JSON(http.StatusCreated, gin.H{"message": "Tugas berhasil ditambahkan", "data": assignment})
}

func UpdateAssignment(c *gin.Context) {
	id := c.Param("id")
	var assignment models.Assignment

	if err := config.DB.First(&assignment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tugas tidak ditemukan"})
		return
	}
	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil || assignment.TeacherID != teacher.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak mengubah tugas ini"})
			return
		}
	}

	var input AssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil || !teacherOwnsClassSubject(teacher.ID, input.ClassSubjectID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak ditugaskan mengajar kelas dan mata pelajaran ini"})
			return
		}
		input.TeacherID = teacher.ID
	} else if input.TeacherID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guru wajib dipilih"})
		return
	}

	assignment.ClassSubjectID = input.ClassSubjectID
	assignment.TeacherID = input.TeacherID
	assignment.Title = input.Title
	assignment.Description = input.Description
	assignment.DueDate = input.DueDate
	
	if input.MaxScore != 0 {
		assignment.MaxScore = input.MaxScore
	}

	if err := config.DB.Save(&assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui tugas"})
		return
	}

	config.DB.Preload("Teacher").First(&assignment, assignment.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Tugas berhasil diperbarui", "data": assignment})
}

func DeleteAssignment(c *gin.Context) {
	id := c.Param("id")
	var assignment models.Assignment
	if err := config.DB.First(&assignment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tugas tidak ditemukan"})
		return
	}
	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil || assignment.TeacherID != teacher.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak menghapus tugas ini"})
			return
		}
	}
	if err := config.DB.Delete(&assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus tugas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tugas berhasil dihapus"})
}
