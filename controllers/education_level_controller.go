package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
)

func GetEducationLevels(c *gin.Context) {
	var levels []models.EducationLevel
	if err := config.DB.Order("id asc").Find(&levels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data jenjang pendidikan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil mengambil data jenjang pendidikan", "data": levels})
}