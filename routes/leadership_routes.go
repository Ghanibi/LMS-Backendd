package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func LeadershipRoutes(router *gin.Engine) {
	group := router.Group("/api/leadership")
	group.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("CURRICULUM", "KURIKULUM", "PRINCIPAL", "KEPSEK", "KEPALA_SEKOLAH"))
	group.GET("/overview", controllers.GetLeadershipOverview)
}
