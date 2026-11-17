package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

func GetDashboardStats(c *gin.Context) {
	counts := map[string]int64{}
	modelsToCount := map[string]any{
		"teachers":    &models.Teacher{},
		"students":    &models.Student{},
		"subjects":    &models.Subject{},
		"assignments": &models.Assignment{},
		"exams":       &models.Exam{},
	}
	for name, model := range modelsToCount {
		var count int64
		if err := config.DB.Model(model).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung statistik " + name})
			return
		}
		counts[name] = count
	}
	c.JSON(http.StatusOK, gin.H{"data": counts})
}
