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
		group.POST("/:type/:id", controllers.UploadAttachments)
		group.GET("/:type/:id", controllers.GetAttachments)
	}
	files := router.Group("/api/attachments/file")
	files.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("ADMIN", "TEACHER", "STUDENT", "SISWA"))
	files.GET("/:attachmentID", controllers.GetAttachmentFile)
}
