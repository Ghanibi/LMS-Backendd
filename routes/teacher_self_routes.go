package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

// TeacherSelfRoutes: endpoint "milik sendiri" khusus untuk guru (bukan CRUD data guru oleh admin).
func TeacherSelfRoutes(router *gin.Engine) {
	group := router.Group("/api/teacher")
	group.Use(middleware.AuthMiddleware())
	group.Use(middleware.RoleMiddleware("TEACHER"))
	{
		group.GET("/me/classes", controllers.GetMyClasses)
	}
}