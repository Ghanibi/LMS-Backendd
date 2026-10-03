package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func ClassroomRoutes(router *gin.Engine) {
	group := router.Group("/api/classrooms")
	group.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("ADMIN", "TEACHER", "STUDENT", "SISWA"))
	{
		group.GET("/mine", controllers.GetMyClassrooms)
		group.GET("/:id/students", controllers.GetClassroomStudents)
		group.GET("/:id/announcements", controllers.GetClassroomAnnouncements)
		group.POST("/:id/announcements", controllers.CreateClassroomAnnouncement)
		group.GET("/:id/discussions", controllers.GetClassroomDiscussions)
		group.POST("/:id/discussions", controllers.CreateClassroomDiscussion)
		group.GET("/:id/attendance", controllers.GetClassroomAttendance)
		group.POST("/:id/attendance", controllers.SaveClassroomAttendance)
		group.GET("/:id/events", controllers.GetClassroomEvents)
		group.POST("/:id/events", controllers.CreateClassroomEvent)
		group.GET("/:id/notes", controllers.GetHomeroomNotes)
		group.POST("/:id/notes", controllers.CreateHomeroomNote)
	}
}
