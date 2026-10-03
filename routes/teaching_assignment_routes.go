package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func TeachingAssignmentRoutes(router *gin.Engine) {
	group := router.Group("/api/teaching-assignments")
	group.Use(middleware.AuthMiddleware())
	group.Use(middleware.RoleMiddleware("ADMIN"))
	{
		group.GET("", controllers.GetTeachingAssignments)
		group.POST("", controllers.CreateTeachingAssignment)
		group.DELETE("/:id", controllers.DeleteTeachingAssignment)
	}
}