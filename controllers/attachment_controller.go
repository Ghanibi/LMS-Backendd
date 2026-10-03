package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

var allowedAttachmentTypes = map[string]bool{
	"assignments":   true,
	"materials":     true,
	"subjects":      true,
	"students":      true,
	"teachers":      true,
	"classes":       true,
	"calendars":     true,
	"announcements": true,
}

var allowedAttachmentExtensions = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".ppt": true, ".pptx": true,
	".xls": true, ".xlsx": true, ".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".webp": true, ".txt": true, ".zip": true,
	".csv": true, ".html": true, ".css": true, ".js": true, ".jsx": true,
	".ts": true, ".tsx": true, ".json": true, ".go": true, ".py": true, ".java": true,
}

func attachmentAccessAllowed(c *gin.Context, resourceType string, resourceID uint) bool {
	if currentRole(c) == "ADMIN" {
		return true
	}
	if currentRole(c) != "TEACHER" || (resourceType != "assignments" && resourceType != "materials") {
		return false
	}
	teacher, err := currentTeacher(c)
	if err != nil {
		return false
	}
	if resourceType == "assignments" {
		var assignment models.Assignment
		return config.DB.Select("id").Where("id = ? AND teacher_id = ?", resourceID, teacher.ID).First(&assignment).Error == nil
	}
	var material models.Material
	return config.DB.Select("id").Where("id = ? AND teacher_id = ?", resourceID, teacher.ID).First(&material).Error == nil
}

func parseAttachmentTarget(c *gin.Context) (string, uint, bool) {
	resourceType := strings.ToLower(c.Param("type"))
	resourceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if !allowedAttachmentTypes[resourceType] || err != nil || resourceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jenis data atau ID lampiran tidak valid"})
		return "", 0, false
	}
	if !attachmentAccessAllowed(c, resourceType, uint(resourceID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak mengelola lampiran untuk data ini"})
		return "", 0, false
	}
	return resourceType, uint(resourceID), true
}

func UploadAttachments(c *gin.Context) {
	resourceType, resourceID, ok := parseAttachmentTarget(c)
	if !ok {
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih file yang akan diunggah"})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih minimal satu file"})
		return
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Folder upload tidak dapat dibuat"})
		return
	}

	created := make([]models.Attachment, 0, len(files))
	for _, file := range files {
		if file.Size <= 0 || file.Size > 20*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran setiap file harus kurang dari 20 MB"})
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedAttachmentExtensions[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung: " + ext})
			return
		}
		randomBytes := make([]byte, 16)
		if _, err := rand.Read(randomBytes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyiapkan nama file"})
			return
		}
		fileName := hex.EncodeToString(randomBytes) + ext
		filePath := filepath.Join("uploads", fileName)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
			return
		}

		attachment := models.Attachment{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			FileName:     filepath.Base(file.Filename),
			FileURL:      "/uploads/" + fileName,
			UploadedBy:   currentUserID(c),
		}
		if err := config.DB.Create(&attachment).Error; err != nil {
			_ = os.Remove(filePath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data lampiran"})
			return
		}
		attachment.FileURL = "/api/attachments/file/" + strconv.FormatUint(uint64(attachment.ID), 10)
		created = append(created, attachment)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "File berhasil diunggah", "data": created})
}

func GetAttachments(c *gin.Context) {
	resourceType, resourceID, ok := parseAttachmentTarget(c)
	if !ok {
		return
	}
	var attachments []models.Attachment
	if err := config.DB.Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Order("created_at ASC").Find(&attachments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar lampiran"})
		return
	}
	for index := range attachments {
		attachments[index].FileURL = "/api/attachments/file/" + strconv.FormatUint(uint64(attachments[index].ID), 10)
	}
	c.JSON(http.StatusOK, gin.H{"data": attachments})
}

func GetAttachmentFile(c *gin.Context) {
	var attachment models.Attachment
	if err := config.DB.First(&attachment, c.Param("attachmentID")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lampiran tidak ditemukan"})
		return
	}
	if !attachmentAccessAllowed(c, attachment.ResourceType, attachment.ResourceID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak mengakses lampiran ini"})
		return
	}
	storedName := filepath.Base(strings.TrimPrefix(attachment.FileURL, "/uploads/"))
	c.FileAttachment(filepath.Join("uploads", storedName), attachment.FileName)
}
