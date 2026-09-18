package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"school-management/config"
	"school-management/models"
)

type CreateStudentInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	NIS      string `json:"nis" binding:"required"`
	NISN     string `json:"nisn"`
	Gender   string `json:"gender"`
	ClassID  uint   `json:"class_id" binding:"required"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type UpdateStudentInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	NIS      string `json:"nis"`
	NISN     string `json:"nisn"`
	Gender   string `json:"gender"`
	ClassID  uint   `json:"class_id"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

func GetStudents(c *gin.Context) {
	var students []models.Student
	// Mengambil data siswa lengkap dengan data User (akun) dan Class (kelas)
	if err := config.DB.Preload("User").Preload("Class").Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data siswa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data siswa",
		"data":    students,
	})
}

func GetStudentByID(c *gin.Context) {
	id := c.Param("id")
	var student models.Student

	if err := config.DB.Preload("User").Preload("Class").First(&student, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Siswa tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil detail siswa",
		"data":    student,
	})
}

func CreateStudent(c *gin.Context) {
	var input CreateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek ketersediaan email
	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email sudah terdaftar"})
		return
	}

	// Cek apakah ClassID valid
	var class models.Class
	if err := config.DB.First(&class, input.ClassID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kelas tidak ditemukan"})
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
		Role:     "STUDENT",
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun user untuk siswa"})
		return
	}

	// 2. Buat Student
	student := models.Student{
		UserID:  user.ID,
		NIS:     input.NIS,
		NISN:    input.NISN,
		Gender:  input.Gender,
		ClassID: input.ClassID,
		Phone:   input.Phone,
		Address: input.Address,
	}
	if err := tx.Create(&student).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat data siswa"})
		return
	}

	tx.Commit()

	// Ambil data terbaru untuk ditampilkan di response
	config.DB.Preload("User").Preload("Class").First(&student, student.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Siswa berhasil ditambahkan",
		"data":    student,
	})
}

func UpdateStudent(c *gin.Context) {
	id := c.Param("id")
	var student models.Student

	if err := config.DB.First(&student, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Siswa tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}

	var input UpdateStudentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := config.DB.Begin()

	// Update data User
	var user models.User
	if err := tx.First(&user, student.UserID).Error; err == nil {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui akun user siswa"})
			return
		}
	}

	// Cek ketersediaan kelas baru jika diupdate
	if input.ClassID != 0 {
		var class models.Class
		if err := tx.First(&class, input.ClassID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kelas tidak ditemukan"})
			return
		}
		student.ClassID = input.ClassID
	}

	// Update data Student
	if input.NIS != "" {
		student.NIS = input.NIS
	}
	if input.NISN != "" {
		student.NISN = input.NISN
	}
	if input.Gender != "" {
		student.Gender = input.Gender
	}
	if input.Phone != "" {
		student.Phone = input.Phone
	}
	if input.Address != "" {
		student.Address = input.Address
	}

	if err := tx.Save(&student).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui data siswa"})
		return
	}

	tx.Commit()

	config.DB.Preload("User").Preload("Class").First(&student, student.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Data siswa berhasil diperbarui",
		"data":    student,
	})
}

func DeleteStudent(c *gin.Context) {
	id := c.Param("id")
	var student models.Student

	if err := config.DB.First(&student, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Siswa tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server"})
		return
	}

	tx := config.DB.Begin()

	if err := tx.Delete(&student).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus data siswa"})
		return
	}

	if err := tx.Delete(&models.User{}, student.UserID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus akun user siswa"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Siswa dan akun terkait berhasil dihapus",
	})
}