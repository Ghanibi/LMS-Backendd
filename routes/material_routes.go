package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func MaterialRoutes(router *gin.Engine) {
	materialGroup := router.Group("/api/materials")

	materialGroup.Use(middleware.AuthMiddleware())
	materialGroup.Use(middleware.RoleMiddleware("ADMIN", "TEACHER"))
	{
		materialGroup.GET("", controllers.GetMaterials)
		materialGroup.GET("/:id", controllers.GetMaterialByID)
		materialGroup.POST("", controllers.CreateMaterial)
		materialGroup.PUT("/:id", controllers.UpdateMaterial)
		materialGroup.DELETE("/:id", controllers.DeleteMaterial)
	}
}