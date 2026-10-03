package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

func canAccessClassroom(c *gin.Context, classID uint) bool {
	if currentRole(c) == "ADMIN" {
		return true
	}
	if currentRole(c) == "TEACHER" {
		teacher, err := currentTeacher(c)
		if err != nil {
			return false
		}
		var count int64
		config.DB.Model(&models.Class{}).Where("id = ? AND homeroom_teacher_id = ?", classID, teacher.ID).Count(&count)
		return count > 0
	}
	if currentRole(c) == "STUDENT" || currentRole(c) == "SISWA" {
		var count int64
		config.DB.Model(&models.Student{}).Where("user_id = ? AND class_id = ?", currentUserID(c), classID).Count(&count)
		return count > 0
	}
	return false
}

func canManageClassroom(c *gin.Context, classID uint) bool {
	if currentRole(c) == "ADMIN" {
		return true
	}
	if currentRole(c) != "TEACHER" {
		return false
	}
	teacher, err := currentTeacher(c)
	if err != nil {
		return false
	}
	var count int64
	config.DB.Model(&models.Class{}).Where("id = ? AND homeroom_teacher_id = ?", classID, teacher.ID).Count(&count)
	return count > 0
}

func GetMyClassrooms(c *gin.Context) {
	var classes []models.Class
	switch currentRole(c) {
	case "TEACHER":
		teacher, err := currentTeacher(c)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Data guru tidak ditemukan"})
			return
		}
		if err := config.DB.Where("homeroom_teacher_id = ?", teacher.ID).Order("name asc").Find(&classes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil kelas wali"})
			return
		}
	case "STUDENT", "SISWA":
		var student models.Student
		if err := config.DB.Preload("Class").Where("user_id = ?", currentUserID(c)).First(&student).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data siswa tidak ditemukan"})
			return
		}
		classes = []models.Class{student.Class}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Role tidak dapat mengakses Kelas Saya"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": classes})
}

func GetClassroomStudents(c *gin.Context) {
	classID := uint(parseUintParam(c, "id"))
	if classID == 0 || !canAccessClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak melihat siswa kelas ini"})
		return
	}
	var students []models.Student
	if err := config.DB.Preload("User").Where("class_id = ?", classID).Order("nis asc").Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar siswa"})
		return
	}
	result := make([]gin.H, 0, len(students))
	for _, student := range students {
		result = append(result, gin.H{"ID": student.ID, "Name": student.User.Name, "NIS": student.NIS})
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func GetClassroomAnnouncements(c *gin.Context) {
	classID := uint(parseUintParam(c, "id"))
	if classID == 0 || !canAccessClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak melihat pengumuman kelas ini"})
		return
	}
	var items []models.Announcement
	if err := config.DB.Where("class_id = ?", classID).Order("created_at desc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil pengumuman kelas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func CreateClassroomAnnouncement(c *gin.Context) {
	classID := uint(parseUintParam(c, "id"))
	if classID == 0 || !canManageClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya wali kelas yang dapat membuat pengumuman kelas"})
		return
	}
	var input AnnouncementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	item := models.Announcement{
		Title: input.Title, Content: input.Content, TargetRole: "Siswa",
		ClassID: &classID, AuthorID: currentUserID(c), IsPublished: true, PublishedAt: &now,
	}
	if err := config.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat pengumuman kelas"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

type ClassDiscussionInput struct {
	ParentID *uint  `json:"parent_id"`
	Title    string `json:"title"`
	Content  string `json:"content" binding:"required"`
}

func GetClassroomDiscussions(c *gin.Context) {
	classID := uint(parseUintParam(c, "id"))
	if classID == 0 || !canAccessClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak melihat diskusi kelas ini"})
		return
	}
	var posts []models.ClassDiscussion
	if err := config.DB.Preload("Author").Where("class_id = ?", classID).Order("created_at asc").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil diskusi kelas"})
		return
	}
	result := make([]gin.H, 0, len(posts))
	for _, post := range posts {
		result = append(result, classDiscussionJSON(post))
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func CreateClassroomDiscussion(c *gin.Context) {
	classID := uint(parseUintParam(c, "id"))
	if classID == 0 || !canAccessClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak mengirim pesan ke kelas ini"})
		return
	}
	var input ClassDiscussionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ParentID == nil && strings.TrimSpace(input.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Judul diskusi wajib diisi"})
		return
	}
	if input.ParentID != nil {
		var parent models.ClassDiscussion
		if err := config.DB.Where("id = ? AND class_id = ? AND parent_id IS NULL", *input.ParentID, classID).First(&parent).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Topik diskusi tidak ditemukan"})
			return
		}
	}
	post := models.ClassDiscussion{
		ClassID: classID, AuthorID: currentUserID(c), ParentID: input.ParentID,
		Title: strings.TrimSpace(input.Title), Content: strings.TrimSpace(input.Content),
	}
	if err := config.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim pesan diskusi"})
		return
	}
	config.DB.Preload("Author").First(&post, post.ID)
	c.JSON(http.StatusCreated, gin.H{"data": classDiscussionJSON(post)})
}

func classDiscussionJSON(post models.ClassDiscussion) gin.H {
	return gin.H{
		"ID": post.ID, "ClassID": post.ClassID, "AuthorID": post.AuthorID,
		"ParentID": post.ParentID, "Title": post.Title, "Content": post.Content,
		"CreatedAt": post.CreatedAt, "Author": gin.H{"Name": post.Author.Name},
	}
}

func parseUintParam(c *gin.Context, key string) uint {
	id, _ := strconv.ParseUint(c.Param(key), 10, 64)
	return uint(id)
}
