package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func SuperAdminRoutes(router *gin.Engine) {
	accounts := router.Group("/api/super-admin/accounts")
	accounts.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("SUPER_ADMIN"))
	accounts.GET("", controllers.GetSuperAdminAccounts)
	accounts.POST("", controllers.CreateSuperAdminManagedAccount)
	accounts.PUT("/:id", controllers.UpdateSuperAdminManagedAccount)
	accounts.DELETE("/:id", controllers.DeleteSuperAdminManagedAccount)

	bootstrap := router.Group("/api/super-admin-bootstrap")
	bootstrap.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("ADMIN"))
	bootstrap.GET("/status", controllers.GetSuperAdminBootstrapStatus)
	bootstrap.POST("", controllers.BootstrapSuperAdmin)
}
