package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"school-management/config"
	"school-management/models"
)

func currentStudent(c *gin.Context) (*models.Student, error) {
	var student models.Student
	err := config.DB.Preload("User").Preload("Class").Where("user_id = ?", currentUserID(c)).First(&student).Error
	return &student, err
}

func loadStudentAttachments(resourceType string, resourceIDs []uint) map[uint][]models.Attachment {
	result := make(map[uint][]models.Attachment)
	if len(resourceIDs) == 0 {
		return result
	}
	var attachments []models.Attachment
	if err := config.DB.Where("resource_type = ? AND resource_id IN ?", resourceType, resourceIDs).Order("created_at asc").Find(&attachments).Error; err != nil {
		return result
	}
	for _, attachment := range attachments {
		attachment.FileURL = "/api/attachments/file/" + strconv.FormatUint(uint64(attachment.ID), 10)
		result[attachment.ResourceID] = append(result[attachment.ResourceID], attachment)
	}
	return result
}

// GetStudentLearning returns only learning data attached to the logged-in student's class.
func GetStudentLearning(c *gin.Context) {
	student, err := currentStudent(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data siswa tidak ditemukan"})
		return
	}
	var classSubjects []models.ClassSubject
	if err := config.DB.Preload("Subject").Where("class_id = ?", student.ClassID).Find(&classSubjects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil mata pelajaran"})
		return
	}
	var timetable []models.ClassSchedule
	if err := config.DB.Preload("Subject").Preload("Teacher").Preload("Teacher.User").Where("class_id = ?", student.ClassID).Find(&timetable).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil jadwal pelajaran"})
		return
	}
	var materials []models.Material
	var assignments []models.Assignment
	var submissions []models.AssignmentSubmission
	var grades []models.Grade
	base := config.DB.Model(&models.ClassSubject{}).Select("id").Where("class_id = ?", student.ClassID)
	if err := config.DB.Preload("Teacher").Preload("Teacher.User").Preload("ClassSubject").Preload("ClassSubject.Subject").Where("class_subject_id IN (?)", base).Order("created_at desc").Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil materi"})
		return
	}
	base = config.DB.Model(&models.ClassSubject{}).Select("id").Where("class_id = ?", student.ClassID)
	if err := config.DB.Preload("Teacher").Preload("Teacher.User").Preload("ClassSubject").Preload("ClassSubject.Subject").Where("class_subject_id IN (?)", base).Order("due_date asc").Find(&assignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tugas"})
		return
	}
	if err := config.DB.Preload("Assignment").Preload("Assignment.ClassSubject").Preload("Assignment.ClassSubject.Subject").Where("student_id = ?", student.ID).Order("updated_at desc").Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil pengumpulan tugas"})
		return
	}
	if err := config.DB.Where("student_id = ?", student.ID).Order("created_at desc").Find(&grades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil nilai"})
		return
	}
	materialIDs := make([]uint, 0, len(materials))
	for _, material := range materials {
		materialIDs = append(materialIDs, material.ID)
	}
	assignmentIDs := make([]uint, 0, len(assignments))
	for _, assignment := range assignments {
		assignmentIDs = append(assignmentIDs, assignment.ID)
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"student": student, "class_subjects": classSubjects, "timetable": timetable, "materials": materials, "assignments": assignments, "submissions": submissions, "grades": grades, "material_attachments": loadStudentAttachments("materials", materialIDs), "assignment_attachments": loadStudentAttachments("assignments", assignmentIDs)}})
}

func SubmitStudentAssignment(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 11<<20)
	student, err := currentStudent(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data siswa tidak ditemukan"})
		return
	}
	assignmentID := uint(parseUintParam(c, "assignmentID"))
	if assignmentID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tugas tidak valid"})
		return
	}
	var assignment models.Assignment
	if err := config.DB.Joins("JOIN class_subjects ON class_subjects.id = assignments.class_subject_id").Where("assignments.id = ? AND class_subjects.class_id = ?", assignmentID, student.ClassID).First(&assignment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tugas tidak ditemukan untuk kelasmu"})
		return
	}
	text := strings.TrimSpace(c.PostForm("submission_text"))
	file, header, fileErr := c.Request.FormFile("file")
	fileURL := ""
	if fileErr == nil {
		defer file.Close()
		if header.Size > 10*1024*1024 {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Ukuran file maksimal 10 MB"})
			return
		}
		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".pdf": true, ".doc": true, ".docx": true, ".ppt": true, ".pptx": true, ".xls": true, ".xlsx": true, ".zip": true, ".png": true, ".jpg": true, ".jpeg": true, ".txt": true, ".csv": true}
		if !allowed[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung"})
			return
		}
		dir := filepath.Join("uploads", "student-submissions")
		if err := os.MkdirAll(dir, 0750); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
			return
		}
		fileName := fmt.Sprintf("%d-%d-%d%s", student.ID, assignment.ID, time.Now().UnixNano(), ext)
		destination := filepath.Join(dir, fileName)
		if err := c.SaveUploadedFile(header, destination); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
			return
		}
		fileURL = "/api/student/me/submissions/files/" + fileName
	} else if fileErr != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak dapat dibaca"})
		return
	}
	if text == "" && fileURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Isi jawaban atau lampirkan file terlebih dahulu"})
		return
	}
	now := time.Now()
	var submission models.AssignmentSubmission
	err = config.DB.Where("assignment_id = ? AND student_id = ?", assignment.ID, student.ID).First(&submission).Error
	if err == nil {
		submission.SubmissionText = text
		if fileURL != "" {
			submission.FileURL = fileURL
		}
		submission.SubmittedAt = &now
		submission.Status = "SUBMITTED"
		submission.Score = nil
		submission.Feedback = ""
		err = config.DB.Save(&submission).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		submission = models.AssignmentSubmission{AssignmentID: assignment.ID, StudentID: student.ID, SubmissionText: text, FileURL: fileURL, SubmittedAt: &now, Status: "SUBMITTED"}
		err = config.DB.Create(&submission).Error
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca status pengumpulan tugas"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengumpulkan tugas"})
		return
	}
	config.DB.Preload("Assignment").First(&submission, submission.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Tugas berhasil dikumpulkan", "data": submission})
}

func GetStudentSubmissionFile(c *gin.Context) {
	student, studentErr := currentStudent(c)
	fileName := filepath.Base(c.Param("fileName"))
	var submission models.AssignmentSubmission
	err := config.DB.Preload("Assignment").Where("file_url = ?", "/api/student/me/submissions/files/"+fileName).First(&submission).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File tidak ditemukan"})
		return
	}
	allowed := studentErr == nil && submission.StudentID == student.ID
	if !allowed && currentRole(c) == "TEACHER" {
		teacher, teacherErr := currentTeacher(c)
		allowed = teacherErr == nil && submission.Assignment.TeacherID == teacher.ID
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tidak memiliki akses ke file ini"})
		return
	}
	c.File(filepath.Join("uploads", "student-submissions", fileName))
}
