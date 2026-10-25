package controllers

import (
	"net/http"
	"school-management/config"
	"school-management/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Mengganti nama struct dan menambahkan validasi
type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginInput

	// Validasi input JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Mencari user berdasarkan email
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Jika user tidak ditemukan, berikan error yang sama seperti password salah
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
			return
		}
		// Jika ada error lain dari database
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencari user"})
		return
	}

	// Membandingkan password yang diinput dengan hash di database
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	// Membuat JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   tokenString,
		// Menggunakan format yang konsisten untuk data user
		"user": gin.H{
			"ID":    user.ID,
			"Name":  user.Name,
			"Email": user.Email,
			"Role":  user.Role,
		},
	})
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
}

// ChangePassword changes only the authenticated user's own password.
func ChangePassword(c *gin.Context) {
	userID := currentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi pengguna tidak valid"})
		return
	}
	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password baru minimal 8 karakter dan maksimal 72 byte"})
		return
	}
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun tidak ditemukan"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password lama tidak sesuai"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.NewPassword)) == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password baru harus berbeda dari password lama"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password baru gagal diamankan"})
		return
	}
	if err := config.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password gagal diperbarui"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}

// FUNGSI BARU: GetMe
// Fungsi ini untuk mengambil data user yang sedang login dari context.
// Data ini didapatkan dari token JWT yang sudah divalidasi oleh middleware.
func GetMe(c *gin.Context) {
	// Ambil data user dari context yang sudah di-set oleh middleware
	userID, _ := c.Get("user_id")
	name, _ := c.Get("name")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	// Kirim data sebagai response
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"name":    name,
		"email":   email,
		"role":    role,
	})
}
