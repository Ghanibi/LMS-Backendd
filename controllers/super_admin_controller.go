package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"school-management/config"
	"school-management/models"
)

type SuperAdminAccountInput struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateSuperAdminAccountInput struct {
	Name     string `json:"name" binding:"omitempty,min=2"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

func GetSuperAdminAccounts(c *gin.Context) {
	var users []models.User
	if err := config.DB.Where("role = ?", "ADMIN").Order("name asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil akun Admin"})
		return
	}
	for i := range users {
		users[i].Password = ""
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func CreateSuperAdminManagedAccount(c *gin.Context) {
	var input SuperAdminAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	if emailTakenByOther(input.Email, 0) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengamankan password"})
		return
	}
	user := models.User{Name: input.Name, Email: input.Email, Password: string(hashedPassword), Role: "ADMIN"}
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun Admin. Pastikan email belum digunakan."})
		return
	}
	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{"message": "Akun Admin berhasil dibuat", "data": user})
}

func UpdateSuperAdminManagedAccount(c *gin.Context) {
	var user models.User
	if err := config.DB.Where("id = ? AND role = ?", c.Param("id"), "ADMIN").First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun Admin tidak ditemukan"})
		return
	}
	var input UpdateSuperAdminAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Name != "" {
		user.Name = strings.TrimSpace(input.Name)
	}
	if input.Email != "" {
		input.Email = strings.TrimSpace(input.Email)
		if emailTakenByOther(input.Email, user.ID) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
			return
		}
		user.Email = input.Email
	}
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengamankan password baru"})
			return
		}
		user.Password = string(hashedPassword)
	}
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui akun Admin"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"message": "Akun Admin berhasil diperbarui", "data": user})
}

func DeleteSuperAdminManagedAccount(c *gin.Context) {
	result := config.DB.Where("id = ? AND role = ?", c.Param("id"), "ADMIN").Delete(&models.User{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus akun Admin"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun Admin tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Akun Admin berhasil dihapus"})
}

func GetSuperAdminBootstrapStatus(c *gin.Context) {
	if currentRole(c) != "ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bootstrap Super Admin hanya dapat dilakukan oleh Admin awal"})
		return
	}
	var count int64
	if err := config.DB.Model(&models.User{}).Where("role = ?", "SUPER_ADMIN").Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa status Super Admin"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"available": count == 0})
}

func BootstrapSuperAdmin(c *gin.Context) {
	if currentRole(c) != "ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bootstrap Super Admin hanya dapat dilakukan oleh Admin awal"})
		return
	}
	var input SuperAdminAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var count int64
	if err := config.DB.Model(&models.User{}).Where("role = ?", "SUPER_ADMIN").Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa akun Super Admin"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Super Admin sudah dibuat. Minta Super Admin saat ini untuk mengelola akun."})
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	if emailTakenByOther(input.Email, 0) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengamankan password"})
		return
	}
	user := models.User{Name: input.Name, Email: input.Email, Password: string(hashedPassword), Role: "SUPER_ADMIN"}
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat Super Admin. Pastikan email belum digunakan."})
		return
	}
	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{"message": "Super Admin berhasil dibuat", "data": user})
}
