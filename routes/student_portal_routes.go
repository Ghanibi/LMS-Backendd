package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func StudentPortalRoutes(router *gin.Engine) {
	portal := router.Group("/api/student/me")
	portal.Use(middleware.AuthMiddleware())
	portal.Use(middleware.RoleMiddleware("STUDENT", "SISWA"))
	portal.GET("/learning", controllers.GetStudentLearning)
	portal.POST("/submissions/:assignmentID", controllers.SubmitStudentAssignment)

	files := router.Group("/api/student/me/submissions/files")
	files.Use(middleware.AuthMiddleware())
	files.Use(middleware.RoleMiddleware("STUDENT", "SISWA", "TEACHER"))
	files.GET("/:fileName", controllers.GetStudentSubmissionFile)
}
