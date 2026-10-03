package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

func emailTakenByOther(email string, exceptUserID uint) bool {
	var count int64
	q := config.DB.Model(&models.User{}).Where("email = ?", email)
	if exceptUserID != 0 {
		q = q.Where("id <> ?", exceptUserID)
	}
	q.Count(&count)
	return count > 0
}

func nipTakenByOther(nip string, exceptTeacherID uint) bool {
	var count int64
	q := config.DB.Model(&models.Teacher{}).Where("nip = ?", nip)
	if exceptTeacherID != 0 {
		q = q.Where("id <> ?", exceptTeacherID)
	}
	q.Count(&count)
	return count > 0
}

func nisTakenByOther(nis string, exceptStudentID uint) bool {
	var count int64
	q := config.DB.Model(&models.Student{}).Where("nis = ?", nis)
	if exceptStudentID != 0 {
		q = q.Where("id <> ?", exceptStudentID)
	}
	q.Count(&count)
	return count > 0
}

// currentRole mengambil role pengguna yang sedang login dari context JWT.
func currentRole(c *gin.Context) string {
	if v, ok := c.Get("role"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// currentUserID mengambil user_id pengguna yang sedang login dari context JWT.
func currentUserID(c *gin.Context) uint {
	if v, ok := c.Get("user_id"); ok {
		if f, ok := v.(float64); ok {
			return uint(f)
		}
	}
	return 0
}

// currentTeacher mengambil data guru (Teacher) milik pengguna yang sedang login.
func currentTeacher(c *gin.Context) (*models.Teacher, error) {
	userID := currentUserID(c)
	if userID == 0 {
		return nil, errors.New("user tidak terautentikasi")
	}
	var teacher models.Teacher
	if err := config.DB.Where("user_id = ?", userID).First(&teacher).Error; err != nil {
		return nil, err
	}
	return &teacher, nil
}