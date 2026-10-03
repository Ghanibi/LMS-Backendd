package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func AssignmentRoutes(router *gin.Engine) {
	assignmentGroup := router.Group("/api/assignments")

	assignmentGroup.Use(middleware.AuthMiddleware())
	assignmentGroup.Use(middleware.RoleMiddleware("ADMIN", "TEACHER"))
	{
		assignmentGroup.GET("", controllers.GetAssignments)
		assignmentGroup.GET("/:id", controllers.GetAssignmentByID)
		assignmentGroup.POST("", controllers.CreateAssignment)
		assignmentGroup.PUT("/:id", controllers.UpdateAssignment)
		assignmentGroup.DELETE("/:id", controllers.DeleteAssignment)
	}
}