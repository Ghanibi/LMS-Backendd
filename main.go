package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
	"school-management/routes"
	"school-management/seed"
)

func main() {
	// 1. Hubungkan ke Database
	config.ConnectDatabase()

	// 2. AutoMigrate semua model
	log.Println("Menjalankan AutoMigrate...")
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.EducationLevel{},
		&models.Class{},
		&models.Subject{},
		&models.Teacher{},
		&models.Student{},
		&models.ClassSubject{},
		&models.TeacherSubject{},
		&models.Material{},
		&models.Assignment{},
		&models.AssignmentSubmission{},
		&models.Exam{},
		&models.ExamQuestion{},
		&models.ExamAnswer{},
		&models.Grade{},
		&models.Announcement{},
		&models.ClassDiscussion{},
		&models.ClassAttendance{},
		&models.ClassroomEvent{},
		&models.HomeroomNote{},
		&models.Notification{},
		&models.AcademicEvent{},
		&models.Attachment{},
	)
	if err != nil {
		log.Fatalf("Gagal melakukan AutoMigrate: %v", err)
	}
	log.Println("AutoMigrate berhasil.")

	// 3. Jalankan Seeder
	seed.SeedEducationLevels()
	seed.SeedClasses()
	seed.SeedSubjects()
	seed.SeedCalendars()
	seed.SeedUsers()

	// 4. Setup Gin Server
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Daftarkan routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	routes.TeacherRoutes(router)
	routes.TeacherSelfRoutes(router)
	routes.StudentRoutes(router)
	routes.ClassRoutes(router)
	routes.EducationLevelRoutes(router)
	routes.SubjectRoutes(router)
	routes.MaterialRoutes(router)
	routes.AssignmentRoutes(router)
	routes.SubmissionRoutes(router)
	routes.AnnouncementRoutes(router)
	routes.ClassroomRoutes(router)
	routes.CalendarRoutes(router)
	routes.TeachingAssignmentRoutes(router)
	routes.AttachmentRoutes(router)

	log.Println("Server berjalan di http://localhost:8080")
	router.Run(":8080")
}
