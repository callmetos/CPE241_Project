package routes

import (
	"car-rental-management/controllers"
	"car-rental-management/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter sets up all the routes for the application
func SetupRouter() *gin.Engine {
	// Set Gin to release mode for production
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Add CORS middleware
	r.Use(CORSMiddleware())

	// Add request logging middleware
	r.Use(RequestLogger())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Car Rental Management API is running",
		})
	})

	api := r.Group("/api")
	{
		// Public routes
		// Registration and Login routes
		api.POST("/register", controllers.RegisterEmployee)
		api.POST("/login", controllers.LoginEmployee)

		// Protected routes requiring JWT authentication
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Cars Management
			protected.GET("/cars", controllers.GetCars)
			protected.POST("/cars", middleware.RoleMiddleware("manager", "admin"), controllers.AddCar)

			// Rentals Management
			protected.GET("/rentals", middleware.RoleMiddleware("manager", "admin"), controllers.GetRentals)
			protected.POST("/rentals", middleware.RoleMiddleware("customer", "manager", "admin"), controllers.CreateRental)

			// Payments Management
			protected.GET("/payments", middleware.RoleMiddleware("manager", "admin"), controllers.GetPayments)
			protected.POST("/payments", middleware.RoleMiddleware("customer", "manager", "admin"), controllers.ProcessPayment)

			// User Management (Admin only)
			protected.GET("/users", middleware.RoleMiddleware("admin"), controllers.GetUsers)

			// Customer Management
			protected.GET("/customers", middleware.RoleMiddleware("manager", "admin"), controllers.GetCustomers)
		}
	}

	// Handle 404 Not Found
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Endpoint not found",
		})
	})

	log.Println("✅ Routes configured successfully!")
	return r
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger logs all incoming requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log the request
		log.Printf("📝 %s | %s | %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

		// Process the request
		c.Next()

		// Log the response status
		log.Printf("📊 %s | %s | %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
	}
}
