package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func AnnouncementRoutes(router *gin.Engine) {
	group := router.Group("/api/announcements")

	group.Use(middleware.AuthMiddleware())
	group.Use(middleware.RoleMiddleware("ADMIN"))
	{
		group.GET("", controllers.GetAnnouncements)
		group.POST("", controllers.CreateAnnouncement)
		group.PUT("/:id", controllers.UpdateAnnouncement)
		group.DELETE("/:id", controllers.DeleteAnnouncement)
	}
}