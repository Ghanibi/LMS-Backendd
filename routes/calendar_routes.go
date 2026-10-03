package routes

import (
	"github.com/gin-gonic/gin"
	"school-management/controllers"
	"school-management/middleware"
)

func CalendarRoutes(router *gin.Engine) {
	group := router.Group("/api/calendars")

	group.Use(middleware.AuthMiddleware())
	group.Use(middleware.RoleMiddleware("ADMIN"))
	{
		group.GET("", controllers.GetCalendars)
		group.POST("", controllers.CreateCalendar)
		group.PUT("/:id", controllers.UpdateCalendar)
		group.DELETE("/:id", controllers.DeleteCalendar)
	}
}