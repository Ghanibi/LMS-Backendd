package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"school-management/config"
	"school-management/models"
)

type SubjectInput struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

func GetSubjects(c *gin.Context) {
	var subjects []models.Subject
	if err := config.DB.Find(&subjects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data mata pelajaran"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data mata pelajaran", "data": subjects})
}

func GetSubjectByID(c *gin.Context) {
	id := c.Param("id")
	var subject models.Subject
	if err := config.DB.First(&subject, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Mata pelajaran tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil detail mata pelajaran", "data": subject})
}

func CreateSubject(c *gin.Context) {
	var input SubjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subject := models.Subject{
		Name:        input.Name,
		Code:        input.Code,
		Description: input.Description,
	}

	if err := config.DB.Create(&subject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat mata pelajaran, pastikan kode unik"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Mata pelajaran berhasil ditambahkan", "data": subject})
}

func UpdateSubject(c *gin.Context) {
	id := c.Param("id")
	var subject models.Subject

	if err := config.DB.First(&subject, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mata pelajaran tidak ditemukan"})
		return
	}

	var input SubjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subject.Name = input.Name
	subject.Code = input.Code
	subject.Description = input.Description

	if err := config.DB.Save(&subject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui mata pelajaran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mata pelajaran berhasil diperbarui", "data": subject})
}

func DeleteSubject(c *gin.Context) {
	id := c.Param("id")
	var subject models.Subject
	if err := config.DB.First(&subject, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mata pelajaran tidak ditemukan"})
		return
	}
	if err := config.DB.Delete(&subject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus mata pelajaran"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mata pelajaran berhasil dihapus"})
}