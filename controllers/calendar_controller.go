package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

type CalendarInput struct {
	Title       string `json:"title" binding:"required"`
	Date        string `json:"date" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// Bentuk response disesuaikan dengan yang dibaca frontend (ID, Title, Date, Description, Category)
type calendarResponse struct {
	ID          uint   `json:"ID"`
	Title       string `json:"Title"`
	Date        string `json:"Date"`
	Description string `json:"Description"`
	Category    string `json:"Category"`
}

func toCalendarResponse(e models.AcademicEvent) calendarResponse {
	return calendarResponse{
		ID:          e.ID,
		Title:       e.Title,
		Date:        e.StartDate.Format("2006-01-02"),
		Description: e.Description,
		Category:    e.Type,
	}
}

func parseCalendarDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

func GetCalendars(c *gin.Context) {
	var events []models.AcademicEvent
	if err := config.DB.Order("start_date asc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kalender"})
		return
	}
	out := make([]calendarResponse, 0, len(events))
	for _, e := range events {
		out = append(out, toCalendarResponse(e))
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data kalender", "data": out})
}

func CreateCalendar(c *gin.Context) {
	var input CalendarInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := parseCalendarDate(input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
		return
	}
	category := input.Category
	if category == "" {
		category = "Akademik"
	}

	var createdBy uint
	if v, ok := c.Get("user_id"); ok {
		if f, ok := v.(float64); ok {
			createdBy = uint(f)
		}
	}

	event := models.AcademicEvent{
		Title:       input.Title,
		Description: input.Description,
		StartDate:   date,
		EndDate:     date,
		Type:        category,
		CreatedBy:   createdBy,
	}
	if err := config.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat agenda"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Agenda berhasil ditambahkan", "data": toCalendarResponse(event)})
}

func UpdateCalendar(c *gin.Context) {
	var event models.AcademicEvent
	if err := config.DB.First(&event, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agenda tidak ditemukan"})
		return
	}

	var input CalendarInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := parseCalendarDate(input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal harus YYYY-MM-DD"})
		return
	}

	event.Title = input.Title
	event.Description = input.Description
	event.StartDate = date
	event.EndDate = date
	if input.Category != "" {
		event.Type = input.Category
	}

	if err := config.DB.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui agenda"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Agenda berhasil diperbarui", "data": toCalendarResponse(event)})
}

func DeleteCalendar(c *gin.Context) {
	var event models.AcademicEvent
	if err := config.DB.First(&event, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agenda tidak ditemukan"})
		return
	}
	if err := config.DB.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus agenda"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Agenda berhasil dihapus"})
}