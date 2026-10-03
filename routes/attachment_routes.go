package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func AttachmentRoutes(router *gin.Engine) {
	group := router.Group("/api/attachments")
	group.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("ADMIN", "TEACHER"))
	{
		group.GET("/file/:attachmentID", controllers.GetAttachmentFile)
		group.POST("/:type/:id", controllers.UploadAttachments)
		group.GET("/:type/:id", controllers.GetAttachments)
	}
}
