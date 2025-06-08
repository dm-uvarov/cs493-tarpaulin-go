package main

import (
	"log"
	"net/http"
	"os"
	//"strings"

	"tarpaulin/api"
	"tarpaulin/auth"
	"tarpaulin/datastore"

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

	// Add middleware to set stores in context
	r.Use(func(c *gin.Context) {
		c.Set("userStore", userStore)
		c.Set("courseStore", courseStore)
		c.Next()
	})

	// Add datastore health check endpoint
	r.GET("/datastore-health", datastoreHealthCheckHandler(userStore))

	// API routes
	// User routes
	userRoutes := r.Group("/users")
	userRoutes.POST("/login", api.LoginHandler(authService))
	userRoutes.GET("/", auth.AuthMiddleware(authService), api.AdminOnlyMiddleware(authService), api.GetAllUsersHandler(userStore))
	userRoutes.GET("/:id", auth.AuthMiddleware(authService), api.GetUserHandler(userStore, courseStore))
	// TODO: Add avatar routes
	// userRoutes.POST("/:id/avatar", auth.AuthMiddleware(authService), api.UploadAvatarHandler)
	// userRoutes.GET("/:id/avatar", auth.AuthMiddleware(authService), api.GetAvatarHandler)
	// userRoutes.DELETE("/:id/avatar", auth.AuthMiddleware(authService), api.DeleteAvatarHandler)

	// Course routes
	courseRoutes := r.Group("/courses")
	// Fix: Add auth middleware before the admin/instructor middleware
	courseRoutes.POST("/", auth.AuthMiddleware(authService), api.AdminOnlyMiddleware(authService), api.CreateCourseHandler)
	courseRoutes.GET("/", api.GetAllCoursesHandler) // Unprotected
	courseRoutes.GET("/:id", api.GetCourseHandler)  // Unprotected
	courseRoutes.PATCH("/:id", auth.AuthMiddleware(authService), api.AdminOnlyMiddleware(authService), api.UpdateCourseHandler)
	courseRoutes.DELETE("/:id", auth.AuthMiddleware(authService), api.AdminOnlyMiddleware(authService), api.DeleteCourseHandler)
	courseRoutes.PATCH("/:id/students", auth.AuthMiddleware(authService), api.CourseInstructorMiddleware(authService), api.UpdateEnrollmentHandler)
	courseRoutes.GET("/:id/students", auth.AuthMiddleware(authService), api.CourseInstructorMiddleware(authService), api.GetEnrollmentHandler)

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

// Inside main() function, after other routes
