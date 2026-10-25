package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func ClassRoutes(router *gin.Engine) {
	classGroup := router.Group("/api/classes")

	classGroup.Use(middleware.AuthMiddleware())
	classGroup.Use(middleware.RoleMiddleware("ADMIN"))
	{
		classGroup.GET("", controllers.GetClasses)
		classGroup.GET("/:id", controllers.GetClassByID)
		classGroup.GET("/:id/schedule", controllers.GetClassSchedule)
		classGroup.PUT("/:id/schedule", controllers.ReplaceClassSchedule)
		classGroup.POST("", controllers.CreateClass)
		classGroup.PUT("/:id", controllers.UpdateClass)
		classGroup.DELETE("/:id", controllers.DeleteClass)
	}
}
