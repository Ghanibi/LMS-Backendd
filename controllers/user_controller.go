package controllers

import (
	"net/http"
	"school-management/config"
	"school-management/models"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var leadershipAccountRoles = map[string]bool{"CURRICULUM": true, "PRINCIPAL": true}

type LeadershipAccountInput struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

type UpdateLeadershipAccountInput struct {
	Name     string `json:"name" binding:"omitempty,min=2"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
	Role     string `json:"role"`
}

func leadershipAccountRole(value string) (string, bool) {
	role := strings.ToUpper(strings.TrimSpace(value))
	return role, leadershipAccountRoles[role]
}

func GetLeadershipAccounts(c *gin.Context) {
	var users []models.User
	if err := config.DB.Where("role IN ?", []string{"CURRICULUM", "PRINCIPAL"}).Order("role asc, name asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil akun pimpinan dan kurikulum"})
		return
	}
	for i := range users {
		users[i].Password = ""
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func CreateLeadershipAccount(c *gin.Context) {
	var input LeadershipAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, valid := leadershipAccountRole(input.Role)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role akun harus Kurikulum atau Kepala Sekolah"})
		return
	}
	if emailTakenByOther(input.Email, 0) {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengamankan password"})
		return
	}
	user := models.User{Name: strings.TrimSpace(input.Name), Email: strings.TrimSpace(input.Email), Password: string(hashedPassword), Role: role}
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun. Pastikan email belum digunakan."})
		return
	}
	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{"message": "Akun berhasil dibuat", "data": user})
}

func UpdateLeadershipAccount(c *gin.Context) {
	var user models.User
	if err := config.DB.Where("id = ? AND role IN ?", c.Param("id"), []string{"CURRICULUM", "PRINCIPAL"}).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun pimpinan atau kurikulum tidak ditemukan"})
		return
	}
	var input UpdateLeadershipAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Role != "" {
		role, valid := leadershipAccountRole(input.Role)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role akun harus Kurikulum atau Kepala Sekolah"})
			return
		}
		user.Role = role
	}
	if input.Email != "" {
		input.Email = strings.TrimSpace(input.Email)
		if emailTakenByOther(input.Email, user.ID) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
			return
		}
		user.Email = input.Email
	}
	if input.Name != "" {
		user.Name = strings.TrimSpace(input.Name)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui akun"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"message": "Akun berhasil diperbarui", "data": user})
}

func DeleteLeadershipAccount(c *gin.Context) {
	result := config.DB.Where("id = ? AND role IN ?", c.Param("id"), []string{"CURRICULUM", "PRINCIPAL"}).Delete(&models.User{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus akun"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun pimpinan atau kurikulum tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Akun berhasil dihapus"})
}

// CreateUserInput adalah struct untuk validasi input saat membuat user baru.
type CreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

// UpdateUserInput adalah struct untuk validasi input saat memperbarui user.
type UpdateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email,email"`
	Role  string `json:"role"`
}

// CreateUser: Membuat user baru (hanya Admin)
// POST /api/users
func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.EqualFold(strings.TrimSpace(input.Role), "SUPER_ADMIN") || strings.EqualFold(strings.TrimSpace(input.Role), "ADMIN") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun Admin dan Super Admin hanya dapat dikelola melalui menu akun khusus"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal melakukan hash password"})
		return
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     input.Role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat user: " + err.Error()})
		return
	}

	// Kosongkan password sebelum dikirim sebagai response
	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

// GetUsers: Mendapatkan semua user (hanya Admin)
// GET /api/users
func GetUsers(c *gin.Context) {
	var users []models.User
	query := config.DB
	if currentRole(c) != "SUPER_ADMIN" {
		query = query.Where("role <> ?", "SUPER_ADMIN")
	}
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data users"})
		return
	}

	// Praktik yang baik untuk tidak mengirim hash password ke client
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

// GetUserByID: Mendapatkan satu user berdasarkan ID (hanya Admin)
// GET /api/users/:id
func GetUserByID(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	if user.Role == "SUPER_ADMIN" && currentRole(c) != "SUPER_ADMIN" {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	user.Password = ""

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// UpdateUser: Memperbarui data user (hanya Admin)
// PUT /api/users/:id
func UpdateUser(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	if strings.EqualFold(user.Role, "SUPER_ADMIN") || strings.EqualFold(user.Role, "ADMIN") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun Admin dan Super Admin hanya dapat dikelola melalui menu akun khusus"})
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.EqualFold(strings.TrimSpace(input.Role), "SUPER_ADMIN") || strings.EqualFold(strings.TrimSpace(input.Role), "ADMIN") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin tidak dapat menetapkan role Admin atau Super Admin melalui endpoint umum"})
		return
	}

	// Update hanya field yang diisi
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Role != "" {
		user.Role = input.Role
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui user"})
		return
	}

	user.Password = ""

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// DeleteUser: Menghapus user (hanya Admin)
// DELETE /api/users/:id
func DeleteUser(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	if strings.EqualFold(user.Role, "SUPER_ADMIN") || strings.EqualFold(user.Role, "ADMIN") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun Admin dan Super Admin hanya dapat dikelola melalui menu akun khusus"})
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil dihapus"})
}
