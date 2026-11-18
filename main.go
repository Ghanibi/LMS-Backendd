package main

import (
	"log"
	"os"
	"sync/atomic"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"school-management/config"
	"school-management/models"
	"school-management/routes"
	"school-management/seed"
)

var databaseReady atomic.Bool
var databaseStatus atomic.Value

func main() {
	databaseStatus.Store("starting")
	// Start listening before database initialization so Vercel's startup probe
	// can reach the process while TiDB migrations run.
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{envOrDefault("FRONTEND_URL", "http://localhost:5173")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}
		if !databaseReady.Load() {
			c.AbortWithStatusJSON(503, gin.H{"error": "Server sedang menyiapkan database. Coba lagi sebentar."})
			return
		}
		c.Next()
	})
	router.GET("/healthz", func(c *gin.Context) {
		if databaseReady.Load() {
			c.JSON(200, gin.H{"status": "ready"})
			return
		}
		stage, _ := databaseStatus.Load().(string)
		c.JSON(503, gin.H{"status": "initializing", "stage": stage})
	})

	// Daftarkan routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	routes.DashboardRoutes(router)
	routes.SuperAdminRoutes(router)
	routes.TeacherRoutes(router)
	routes.TeacherSelfRoutes(router)
	routes.StudentRoutes(router)
	routes.StudentPortalRoutes(router)
	routes.LeadershipRoutes(router)
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

	port := envOrDefault("PORT", "8080")
	log.Printf("Server berjalan di port %s", port)
	go initializeDatabase(envOrDefault("RUN_DB_INIT", "true") == "true")
	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}

func initializeDatabase(runMigrationsAndSeeds bool) {
	databaseStatus.Store("connecting_to_database")
	log.Println("Menghubungkan ke database...")
	if err := config.ConnectDatabase(); err != nil {
		databaseStatus.Store("database_connection_failed")
		log.Printf("Gagal terhubung ke database: %v", err)
		return
	}
	databaseStatus.Store("database_connected")

	if runMigrationsAndSeeds {
		databaseStatus.Store("running_migrations")
		log.Println("Menjalankan AutoMigrate...")
		err := config.DB.AutoMigrate(
			&models.User{},
			&models.EducationLevel{},
			&models.Class{},
			&models.Subject{},
			&models.Teacher{},
			&models.Student{},
			&models.ClassSubject{},
			&models.ClassSchedule{},
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

		databaseStatus.Store("running_seeders")
		seed.SeedEducationLevels()
		seed.SeedClasses()
		seed.SeedSubjects()
		seed.SeedCalendars()
		seed.SeedUsers()
		seed.SeedClassScheduleSource()
		seed.ResolveUnambiguousScheduleSubjects()
	} else {
		log.Println("RUN_DB_INIT=false; melewati AutoMigrate dan seeder.")
	}

	databaseStatus.Store("ready")
	databaseReady.Store(true)
	log.Println("Database siap menerima request.")
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
