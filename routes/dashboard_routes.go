package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func DashboardRoutes(router *gin.Engine) {
	group := router.Group("/api/dashboard")
	group.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("ADMIN"))
	group.GET("/stats", controllers.GetDashboardStats)
}
