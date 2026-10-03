package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"school-management/config"
	"school-management/models"
)

func classroomDate(c *gin.Context) (time.Time, error) {
	value := c.Query("date")
	if value == "" {
		value = time.Now().Format("2006-01-02")
	}
	return time.ParseInLocation("2006-01-02", value, time.Local)
}

func validAttendanceStatus(status string) bool {
	switch status {
	case "Hadir", "Izin", "Sakit", "Alpa":
		return true
	default:
		return false
	}
}

type attendanceInput struct {
	StudentID uint   `json:"student_id" binding:"required"`
	Status    string `json:"status" binding:"required"`
	Note      string `json:"note"`
}

type attendanceBatchInput struct {
	Date    string            `json:"date" binding:"required"`
	Records []attendanceInput `json:"records" binding:"required,min=1"`
}

func GetClassroomAttendance(c *gin.Context) {
	classID := parseUintParam(c, "id")
	if classID == 0 || !canAccessClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak melihat presensi kelas ini"})
		return
	}
	date, err := classroomDate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
		return
	}

	query := config.DB.Where("class_id = ? AND date = ?", classID, date)
	if currentRole(c) == "STUDENT" || currentRole(c) == "SISWA" {
		var student models.Student
		if err := config.DB.Where("user_id = ? AND class_id = ?", currentUserID(c), classID).First(&student).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Data siswa kelas tidak ditemukan"})
			return
		}
		query = query.Where("student_id = ?", student.ID)
	}
	var rows []models.ClassAttendance
	if err := query.Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil presensi kelas"})
		return
	}
	byStudent := make(map[uint]models.ClassAttendance, len(rows))
	for _, row := range rows {
		byStudent[row.StudentID] = row
	}

	var students []models.Student
	studentQuery := config.DB.Where("class_id = ?", classID)
	if currentRole(c) == "STUDENT" || currentRole(c) == "SISWA" {
		studentQuery = studentQuery.Where("user_id = ?", currentUserID(c))
	}
	if err := studentQuery.Preload("User").Order("nis asc").Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar siswa"})
		return
	}
	result := make([]gin.H, 0, len(students))
	for _, student := range students {
		row, exists := byStudent[student.ID]
		status, note := "Belum diisi", ""
		if exists {
			status, note = row.Status, row.Note
		}
		result = append(result, gin.H{"student_id": student.ID, "student_name": student.User.Name, "nis": student.NIS, "status": status, "note": note})
	}
	c.JSON(http.StatusOK, gin.H{"date": date.Format("2006-01-02"), "data": result})
}

func SaveClassroomAttendance(c *gin.Context) {
	classID := parseUintParam(c, "id")
	if classID == 0 || !canManageClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya wali kelas yang dapat mengatur presensi"})
		return
	}
	var input attendanceBatchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tanggal dan daftar presensi wajib diisi"})
		return
	}
	date, err := time.ParseInLocation("2006-01-02", input.Date, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
		return
	}
	var classStudentIDs []uint
	if err := config.DB.Model(&models.Student{}).Where("class_id = ?", classID).Pluck("id", &classStudentIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memeriksa siswa kelas"})
		return
	}
	validIDs := make(map[uint]bool, len(classStudentIDs))
	for _, id := range classStudentIDs {
		validIDs[id] = true
	}
	userID := currentUserID(c)
	rows := make([]models.ClassAttendance, 0, len(input.Records))
	for _, record := range input.Records {
		if !validIDs[record.StudentID] || !validAttendanceStatus(record.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Siswa atau status presensi tidak valid"})
			return
		}
		rows = append(rows, models.ClassAttendance{ClassID: classID, StudentID: record.StudentID, Date: date, Status: record.Status, Note: strings.TrimSpace(record.Note), TakenBy: userID})
	}
	if err := config.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "class_id"}, {Name: "student_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "note", "taken_by", "updated_at"}),
	}).Create(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan presensi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Presensi berhasil disimpan"})
}

type classroomEventInput struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	StartsAt    time.Time  `json:"starts_at" binding:"required"`
	EndsAt      *time.Time `json:"ends_at"`
}

func GetClassroomEvents(c *gin.Context) {
	classID := parseUintParam(c, "id")
	if classID == 0 || !canAccessClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak berhak melihat agenda kelas ini"})
		return
	}
	var events []models.ClassroomEvent
	if err := config.DB.Where("class_id = ?", classID).Order("starts_at asc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil agenda kelas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

func CreateClassroomEvent(c *gin.Context) {
	classID := parseUintParam(c, "id")
	if classID == 0 || !canManageClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya wali kelas yang dapat membuat agenda"})
		return
	}
	var input classroomEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Judul dan waktu kegiatan wajib diisi"})
		return
	}
	if input.EndsAt != nil && input.EndsAt.Before(input.StartsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Waktu selesai tidak boleh sebelum waktu mulai"})
		return
	}
	event := models.ClassroomEvent{ClassID: classID, Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description), StartsAt: input.StartsAt, EndsAt: input.EndsAt, CreatedBy: currentUserID(c)}
	if err := config.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan agenda kelas"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": event})
}

type homeroomNoteInput struct {
	StudentID uint   `json:"student_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

func GetHomeroomNotes(c *gin.Context) {
	classID := parseUintParam(c, "id")
	if classID == 0 || !canManageClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Catatan pembinaan hanya dapat dilihat wali kelas dan admin"})
		return
	}
	var notes []models.HomeroomNote
	if err := config.DB.Where("class_id = ?", classID).Order("created_at desc").Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil catatan pembinaan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": notes})
}

func CreateHomeroomNote(c *gin.Context) {
	classID := parseUintParam(c, "id")
	if classID == 0 || !canManageClassroom(c, classID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Catatan pembinaan hanya dapat ditulis wali kelas dan admin"})
		return
	}
	var input homeroomNoteInput
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Siswa dan isi catatan wajib diisi"})
		return
	}
	var count int64
	config.DB.Model(&models.Student{}).Where("id = ? AND class_id = ?", input.StudentID, classID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Siswa tidak terdaftar di kelas ini"})
		return
	}
	note := models.HomeroomNote{ClassID: classID, StudentID: input.StudentID, AuthorID: currentUserID(c), Content: strings.TrimSpace(input.Content)}
	if err := config.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan catatan pembinaan"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": note})
}
