package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

type AnnouncementInput struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	TargetRole string `json:"target_role"`
}

func normalizeTarget(t string) string {
	switch t {
	case "Siswa", "Guru":
		return t
	default:
		return "Semua"
	}
}

func GetAnnouncements(c *gin.Context) {
	var items []models.Announcement
	if err := config.DB.Where("class_id IS NULL").Order("created_at desc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengumuman"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data pengumuman", "data": items})
}

func CreateAnnouncement(c *gin.Context) {
	var input AnnouncementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var authorID uint
	if v, ok := c.Get("user_id"); ok {
		if f, ok := v.(float64); ok {
			authorID = uint(f)
		}
	}

	now := time.Now()
	item := models.Announcement{
		Title:       input.Title,
		Content:     input.Content,
		TargetRole:  normalizeTarget(input.TargetRole),
		AuthorID:    authorID,
		IsPublished: true,
		PublishedAt: &now,
	}
	if err := config.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat pengumuman"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Pengumuman berhasil ditambahkan", "data": item})
}

func UpdateAnnouncement(c *gin.Context) {
	var item models.Announcement
	if err := config.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengumuman tidak ditemukan"})
		return
	}

	var input AnnouncementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.Title = input.Title
	item.Content = input.Content
	item.TargetRole = normalizeTarget(input.TargetRole)

	if err := config.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui pengumuman"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pengumuman berhasil diperbarui", "data": item})
}

func DeleteAnnouncement(c *gin.Context) {
	var item models.Announcement
	if err := config.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengumuman tidak ditemukan"})
		return
	}
	if err := config.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pengumuman"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pengumuman berhasil dihapus"})
}
