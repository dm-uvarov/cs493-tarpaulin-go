package main

import (
	"log"
	"net/http"
	"os"

	//"strings"

	"tarpaulin/api"
	"tarpaulin/auth"
	"tarpaulin/datastore"
	"tarpaulin/storage" // Add this line

	//"github.com/form3tech-oss/jwt-go" // Add this for jwt.MapClaims
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	// Initialize router
	r := gin.Default() // Includes logger and recovery middleware

	// Health check endpoint
	r.GET("/health", healthCheckHandler)

	// Initialize Auth0 service
	authService, err := auth.NewAuthService()
	if err != nil {
		log.Fatalf("Failed to initialize Auth0 service: %v", err)
	}

	// Initialize services
	userStore, err := datastore.NewUserStore()
	if err != nil {
		log.Fatalf("Failed to initialize user store: %v", err)
	}

	// Initialize course store
	courseStore, err := datastore.NewCourseStore()
	if err != nil {
		log.Fatalf("Failed to initialize course store: %v", err)
	}

	// Initialize storage service
	storageService, err := storage.NewStorageService()
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}

	// API routes
	// User routes
	userRoutes := r.Group("/users")
	userRoutes.POST("/login", api.LoginHandler(authService))
	userRoutes.GET("/", auth.AuthMiddleware(authService, userStore), api.AdminOnlyMiddleware(authService, userStore), api.GetAllUsersHandler(userStore))
	userRoutes.GET("/:id", auth.AuthMiddleware(authService, userStore), api.GetUserHandler(userStore, courseStore))
	// Avatar routes
	userRoutes.POST("/:id/avatar", auth.AuthMiddleware(authService, userStore), api.UploadAvatarHandler(userStore, storageService))
	userRoutes.GET("/:id/avatar", auth.AuthMiddleware(authService, userStore), api.GetAvatarHandler(userStore, storageService))
	userRoutes.DELETE("/:id/avatar", auth.AuthMiddleware(authService, userStore), api.DeleteAvatarHandler(userStore, storageService))

	// Course routes
	courseRoutes := r.Group("/courses")
	courseRoutes.POST("/", auth.AuthMiddleware(authService, userStore), api.AdminOnlyMiddleware(authService, userStore), api.CreateCourseHandler(courseStore, userStore))
	courseRoutes.GET("/", api.GetAllCoursesHandler(courseStore)) // Unprotected
	courseRoutes.GET("/:id", api.GetCourseHandler(courseStore))   // Updated to pass courseStore
	courseRoutes.PATCH("/:id", auth.AuthMiddleware(authService, userStore), api.AdminOnlyMiddleware(authService, userStore), api.UpdateCourseHandler(courseStore, userStore))
	courseRoutes.DELETE("/:id", auth.AuthMiddleware(authService, userStore), api.AdminOnlyMiddleware(authService, userStore), api.DeleteCourseHandler(courseStore))
	courseRoutes.PATCH("/:id/students", auth.AuthMiddleware(authService, userStore), api.CourseInstructorMiddleware(authService, userStore, courseStore), api.UpdateEnrollmentHandler(courseStore, userStore))
	courseRoutes.GET("/:id/students", auth.AuthMiddleware(authService, userStore), api.CourseInstructorMiddleware(authService, userStore, courseStore), api.GetEnrollmentHandler(courseStore, userStore))

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s\n", port)
	r.Run(":" + port)
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Service is healthy",
	})
}

// Add this new handler function at the end of the file
func datastoreHealthCheckHandler(userStore *datastore.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := userStore.HealthCheck(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Datastore connection failed",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Datastore connection is healthy",
		})
	}
}
