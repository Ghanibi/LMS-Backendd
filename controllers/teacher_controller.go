package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"school-management/config"
	"school-management/models"
)

type CreateTeacherInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password"`
	NIP      string `json:"nip"`
	Gender   string `json:"gender"`
	Subject  string `json:"subject"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type UpdateTeacherInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	NIP      string `json:"nip"`
	Gender   string `json:"gender"`
	Subject  string `json:"subject"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

func GetTeachers(c *gin.Context) {
	var teachers []models.Teacher
	if err := config.DB.Preload("User").Find(&teachers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data guru"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data guru",
		"data":    teachers,
	})
}

func GetTeacherByID(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher

	if err := config.DB.Preload("User").First(&teacher, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Guru tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil detail guru",
		"data":    teacher,
	})
}

func CreateTeacher(c *gin.Context) {
	var input CreateTeacherInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email sudah terdaftar"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	tx := config.DB.Begin()

	// 1. Buat User
	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     "TEACHER",
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun user untuk guru"})
		return
	}

	// 2. Buat Teacher
	teacher := models.Teacher{
		UserID:  user.ID,
		NIP:     input.NIP,
		Gender:  input.Gender,
		Subject: input.Subject,
		Phone:   input.Phone,
		Address: input.Address,
	}
	if err := tx.Create(&teacher).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat data guru"})
		return
	}

	tx.Commit()

	config.DB.Preload("User").First(&teacher, teacher.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Guru berhasil ditambahkan",
		"data":    teacher,
	})
}

func UpdateTeacher(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher

	if err := config.DB.First(&teacher, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Guru tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}

	var input UpdateTeacherInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := config.DB.Begin()

	// Update data User
	var user models.User
	if err := tx.First(&user, teacher.UserID).Error; err == nil {
		if input.Name != "" {
			user.Name = input.Name
		}
		if input.Email != "" {
			user.Email = input.Email
		}
		if input.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
			if err == nil {
				user.Password = string(hashedPassword)
			}
		}
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui akun user guru"})
			return
		}
	}

	// Update data Teacher
	if input.NIP != "" {
		teacher.NIP = input.NIP
	}
	if input.Gender != "" {
		teacher.Gender = input.Gender
	}
	if input.Subject != "" {
		teacher.Subject = input.Subject
	}
	if input.Phone != "" {
		teacher.Phone = input.Phone
	}
	if input.Address != "" {
		teacher.Address = input.Address
	}

	if err := tx.Save(&teacher).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data guru"})
		return
	}

	tx.Commit()

	config.DB.Preload("User").First(&teacher, teacher.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Data guru berhasil diperbarui",
		"data":    teacher,
	})
}

func DeleteTeacher(c *gin.Context) {
	id := c.Param("id")
	var teacher models.Teacher

	if err := config.DB.First(&teacher, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Guru tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}

	tx := config.DB.Begin()

	if err := tx.Delete(&teacher).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus data guru"})
		return
	}

	if err := tx.Delete(&models.User{}, teacher.UserID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus akun user guru"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Guru dan akun terkait berhasil dihapus",
	})
}