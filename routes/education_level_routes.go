package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func EducationLevelRoutes(router *gin.Engine) {
	group := router.Group("/api/education-levels")

	group.Use(middleware.AuthMiddleware())
	group.Use(middleware.RoleMiddleware("ADMIN"))
	{
		group.GET("", controllers.GetEducationLevels)
	}
}