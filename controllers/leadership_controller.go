package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

func GetLeadershipOverview(c *gin.Context) {
	var studentCount, teacherCount, classCount, subjectCount, materialCount, assignmentCount, submissionCount, pendingCount int64
	queries := []struct {
		model  any
		target *int64
	}{
		{&models.Student{}, &studentCount}, {&models.Teacher{}, &teacherCount}, {&models.Class{}, &classCount},
		{&models.Subject{}, &subjectCount}, {&models.Material{}, &materialCount}, {&models.Assignment{}, &assignmentCount}, {&models.AssignmentSubmission{}, &submissionCount},
	}
	for _, query := range queries {
		if err := config.DB.Model(query.model).Count(query.target).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung data sekolah"})
			return
		}
	}
	if err := config.DB.Model(&models.AssignmentSubmission{}).Where("status <> ?", "GRADED").Count(&pendingCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung tugas yang menunggu penilaian"})
		return
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	var attendanceRows []struct {
		Status string
		Count  int64
	}
	if err := config.DB.Model(&models.ClassAttendance{}).Select("status, COUNT(*) as count").Where("date >= ? AND date < ?", monthStart, monthEnd).Group("status").Scan(&attendanceRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil ringkasan presensi"})
		return
	}
	attendanceSummary := map[string]int64{"Hadir": 0, "Izin": 0, "Sakit": 0, "Alpa": 0}
	var recorded int64
	for _, row := range attendanceRows {
		attendanceSummary[row.Status] = row.Count
		recorded += row.Count
	}
	presentRate := 0.0
	if recorded > 0 {
		presentRate = float64(attendanceSummary["Hadir"]) * 100 / float64(recorded)
	}

	var classes []models.Class
	if err := config.DB.Preload("EducationLevel").Preload("HomeroomTeacher").Preload("HomeroomTeacher.User").Find(&classes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil ringkasan kelas"})
		return
	}
	type classSummary struct {
		ID              uint
		Name            string
		Grade           int
		Major           string
		Students        int64
		HomeroomTeacher string
	}
	classRows := make([]classSummary, 0, len(classes))
	for _, class := range classes {
		var count int64
		if err := config.DB.Model(&models.Student{}).Where("class_id = ?", class.ID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung siswa per kelas"})
			return
		}
		teacherName := "Belum ditetapkan"
		if class.HomeroomTeacher != nil && class.HomeroomTeacher.User.Name != "" {
			teacherName = class.HomeroomTeacher.User.Name
		}
		classRows = append(classRows, classSummary{ID: class.ID, Name: class.Name, Grade: class.Grade, Major: class.Major, Students: count, HomeroomTeacher: teacherName})
	}

	var classSubjects []models.ClassSubject
	if err := config.DB.Preload("Class").Preload("Subject").Order("class_id asc").Find(&classSubjects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil peta mata pelajaran"})
		return
	}
	var teachingAssignments []models.TeacherSubject
	if err := config.DB.Preload("Teacher").Preload("Teacher.User").Preload("Subject").Preload("Class").Find(&teachingAssignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil penugasan pengajar"})
		return
	}
	var materials []models.Material
	if err := config.DB.Preload("ClassSubject").Preload("ClassSubject.Class").Preload("ClassSubject.Subject").Preload("Teacher").Preload("Teacher.User").Order("created_at desc").Limit(30).Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil ringkasan materi"})
		return
	}
	var assignments []models.Assignment
	if err := config.DB.Preload("ClassSubject").Preload("ClassSubject.Class").Preload("ClassSubject.Subject").Preload("Teacher").Preload("Teacher.User").Order("due_date asc").Limit(30).Find(&assignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil ringkasan tugas"})
		return
	}
	var schoolEvents []models.AcademicEvent
	if err := config.DB.Where("start_date >= ?", now).Order("start_date asc").Limit(8).Find(&schoolEvents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil agenda akademik"})
		return
	}
	var classEvents []models.ClassroomEvent
	if err := config.DB.Where("starts_at >= ?", now).Order("starts_at asc").Limit(8).Find(&classEvents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil agenda kelas"})
		return
	}
	var announcements []models.Announcement
	if err := config.DB.Where("class_id IS NULL AND is_published = ?", true).Order("created_at desc").Limit(6).Find(&announcements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil pengumuman sekolah"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"stats":      gin.H{"students": studentCount, "teachers": teacherCount, "classes": classCount, "subjects": subjectCount, "materials": materialCount, "assignments": assignmentCount, "submissions": submissionCount, "pending_reviews": pendingCount, "attendance_rate": presentRate, "attendance_recorded": recorded},
		"attendance": attendanceSummary, "classes": classRows, "class_subjects": classSubjects, "teacher_assignments": teachingAssignments,
		"materials": materials, "assignments": assignments, "academic_events": schoolEvents, "class_events": classEvents, "announcements": announcements,
	}})
}
