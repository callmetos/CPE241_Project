package router // <--- ตรวจสอบว่าเป็น package router

import (
	"car-rental-management/internal/handlers"
	"car-rental-management/internal/middleware" // <--- เพิ่ม import fmt
	"log"
	"net/http"
	"os"

	// <--- เพิ่ม import time
	"github.com/gin-gonic/gin"
)

// SetupRouter sets up all the routes for the application
func SetupRouter() *gin.Engine {
	// gin.SetMode(gin.ReleaseMode) // Uncomment for production

	r := gin.Default()

	// --- Global Middleware ---
	r.Use(CORSMiddleware())           // Apply CORS first
	r.Use(middleware.RequestLogger()) // <--- เรียกใช้ RequestLogger ที่เพิ่มเข้ามา

	r.LoadHTMLGlob("templates/*.html")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	api := r.Group("/api")
	{
		// --- Public: Authentication ---
		api.POST("/employee/register", handlers.RegisterEmployee)
		api.POST("/employee/login", handlers.LoginEmployee)
		api.POST("/customer/register", handlers.RegisterCustomer)
		api.POST("/customer/login", handlers.LoginCustomer)

		// --- Public: Information ---
		api.GET("/cars", handlers.GetCars)
		api.GET("/cars/:id", handlers.GetCarByID)
		api.GET("/cars/:id/reviews", handlers.GetCarReviews) // <--- แก้ไข ใช้ :id
		api.GET("/branches", handlers.GetBranches)
		api.GET("/branches/:id", handlers.GetBranchByID)

		// --- Protected Routes ---
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Route for deleting reviews (accessible by authenticated user, permission checked in handler)
			protected.DELETE("/reviews/:id", handlers.DeleteReview) // <--- ย้ายมาไว้ตรงนี้ และมีแค่ที่เดียว

			// --- Employee Routes ---
			staffAdminManager := protected.Group("/")
			staffAdminManager.Use(middleware.RoleMiddleware("admin", "manager"))
			{
				// Branch Management
				staffAdminManager.POST("/branches", handlers.CreateBranch)
				staffAdminManager.PUT("/branches/:id", handlers.UpdateBranch)
				staffAdminManager.DELETE("/branches/:id", handlers.DeleteBranch)

				// Car Management
				staffAdminManager.POST("/cars", handlers.AddCar)
				staffAdminManager.PUT("/cars/:id", handlers.UpdateCar)
				staffAdminManager.DELETE("/cars/:id", handlers.DeleteCar)

				// Customer Management (By Staff)
				staffAdminManager.GET("/customers", handlers.GetCustomers)
				staffAdminManager.GET("/customers/:id", handlers.GetCustomerByID)
				staffAdminManager.PUT("/customers/:id", handlers.UpdateCustomer)
				staffAdminManager.DELETE("/customers/:id", handlers.DeleteCustomer)

				// Rental Management (By Staff)
				staffAdminManager.GET("/rentals", handlers.GetRentals)
				staffAdminManager.GET("/rentals/:id", handlers.GetRentalByIDForStaff)
				staffAdminManager.POST("/rentals/:id/confirm", handlers.ConfirmRental)
				staffAdminManager.POST("/rentals/:id/activate", handlers.ActivateRental)
				staffAdminManager.POST("/rentals/:id/return", handlers.ReturnRental)
				staffAdminManager.POST("/rentals/:id/cancel", handlers.CancelRentalByStaff)

				// Payment Management (By Staff)
				staffAdminManager.GET("/payments", handlers.GetPayments)
				staffAdminManager.GET("/rentals/:id/payments", handlers.GetPaymentsByRental) // <--- แก้ไข ใช้ :id
				staffAdminManager.POST("/rentals/:id/payments", handlers.ProcessPayment)     // <--- แก้ไข ใช้ :id

				// Dashboard
				staffAdminManager.GET("/dashboard", handlers.GetDashboard)

				// Review Management (By Staff)
				// staffAdminManager.DELETE("/reviews/:id", handlers.DeleteReview) // <--- ลบออกจากตรงนี้
			}

			// Admin Only Routes
			adminOnly := protected.Group("/")
			adminOnly.Use(middleware.RoleMiddleware("admin"))
			{
				adminOnly.GET("/users", handlers.GetUsers)
			}

			// --- Customer Routes ---
			customerOnly := protected.Group("/")
			customerOnly.Use(middleware.CustomerRequired())
			{
				// Profile
				customerOnly.GET("/me/profile", handlers.GetMyProfile)
				customerOnly.PUT("/me/profile", handlers.UpdateMyProfile)

				// Rentals
				customerOnly.POST("/rentals", handlers.CreateRental)
				customerOnly.GET("/my/rentals", handlers.GetMyRentals)
				customerOnly.GET("/my/rentals/:id", handlers.GetMyRentalByID)
				customerOnly.POST("/my/rentals/:id/cancel", handlers.CancelMyRental)

				// Reviews
				customerOnly.POST("/my/rentals/:id/review", handlers.SubmitReview) // <--- แก้ไข ใช้ :id
				// customerOnly.DELETE("/reviews/:id", handlers.DeleteReview) // <--- ลบออกจากตรงนี้
			}
		}
	}

	// Handle 404 Not Found
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})

	log.Println("✅ Routes configured successfully!")
	return r
}

// --- CORSMiddleware ---
// (เหมือนเดิม)
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
