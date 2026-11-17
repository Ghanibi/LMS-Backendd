package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func TeacherRoutes(router *gin.Engine) {
	teacherGroup := router.Group("/api/teachers")

	teacherGroup.Use(middleware.AuthMiddleware())
	teacherGroup.Use(middleware.RoleMiddleware("ADMIN"))
	{
		teacherGroup.GET("", controllers.GetTeachers)
		teacherGroup.POST("/import", controllers.ImportTeachersCSV)
		teacherGroup.GET("/:id", controllers.GetTeacherByID)
		teacherGroup.POST("", controllers.CreateTeacher)
		teacherGroup.PUT("/:id", controllers.UpdateTeacher)
		teacherGroup.DELETE("/:id", controllers.DeleteTeacher)
	}
}
