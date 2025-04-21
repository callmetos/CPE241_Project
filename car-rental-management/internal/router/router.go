package router // <--- ตรวจสอบว่าเป็น package router

import (
	"car-rental-management/internal/handlers"
	"car-rental-management/internal/middleware"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupRouter sets up all the routes for the application
func SetupRouter() *gin.Engine {
	// Consider setting Gin mode based on an environment variable
	// gin.SetMode(gin.ReleaseMode) // Example for production

	// --- Future Enhancement: Consider using a structured logger ---
	// e.g., logrus or zap for better log management in production.

	// --- Future Enhancement: Consider Dependency Injection ---
	// Passing dependencies (like DB connection) explicitly instead of using global
	// variables can improve testability and maintainability for larger applications.

	r := gin.Default() // Includes Logger() and Recovery() middleware

	// --- Global Middleware ---
	r.Use(CORSMiddleware())           // Apply CORS first
	r.Use(middleware.RequestLogger()) // Custom request logger

	// Load HTML templates (currently only for 404)
	r.LoadHTMLGlob("templates/*.html")

	// Simple health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API v1 routes
	api := r.Group("/api")
	{
		// --- Public: Authentication ---
		auth := api.Group("/auth") // Group auth routes
		{
			auth.POST("/employee/register", handlers.RegisterEmployee) // Consider admin-only for registration?
			auth.POST("/employee/login", handlers.LoginEmployee)
			auth.POST("/customer/register", handlers.RegisterCustomer)
			auth.POST("/customer/login", handlers.LoginCustomer)
		}

		// --- Public: Browse Information ---
		api.GET("/cars", handlers.GetCars)                   // List/filter cars
		api.GET("/cars/:id", handlers.GetCarByID)            // Get specific car
		api.GET("/cars/:id/reviews", handlers.GetCarReviews) // Get reviews for a specific car (Public)
		api.GET("/branches", handlers.GetBranches)           // List branches
		api.GET("/branches/:id", handlers.GetBranchByID)     // Get specific branch

		// --- Protected Routes (Require Authentication) ---
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware()) // Apply JWT authentication to all routes below
		{
			// --- Authenticated User Actions (Permission checked within handler) ---
			// Any authenticated user (customer or staff) can attempt to delete a review,
			// the handler/service verifies if they *own* it or have *staff* permissions.
			protected.DELETE("/reviews/:id", handlers.DeleteReview)

			// Get review for a specific rental (staff or rental owner only)
			// Requires login, permissions checked in handler.
			protected.GET("/rentals/:id/review", handlers.GetRentalReview)

			// --- Employee Routes (Require Employee Role: Admin or Manager) ---
			staff := protected.Group("/")
			staff.Use(middleware.RoleMiddleware("admin", "manager")) // Only admin/manager allowed
			{
				// Branch Management (Staff)
				staff.POST("/branches", handlers.CreateBranch)
				staff.PUT("/branches/:id", handlers.UpdateBranch)
				staff.DELETE("/branches/:id", handlers.DeleteBranch)

				// Car Management (Staff)
				staff.POST("/cars", handlers.AddCar)
				staff.PUT("/cars/:id", handlers.UpdateCar)
				staff.DELETE("/cars/:id", handlers.DeleteCar)

				// Customer Management (Staff)
				staff.GET("/customers", handlers.GetCustomers)
				staff.GET("/customers/:id", handlers.GetCustomerByID)
				staff.PUT("/customers/:id", handlers.UpdateCustomer) // Staff updates customer
				staff.DELETE("/customers/:id", handlers.DeleteCustomer)

				// Rental Management (Staff)
				staff.GET("/rentals", handlers.GetRentals)                      // Get all rentals
				staff.GET("/rentals/:id", handlers.GetRentalByIDForStaff)       // Get specific rental by staff
				staff.POST("/rentals/:id/confirm", handlers.ConfirmRental)      // Update status
				staff.POST("/rentals/:id/activate", handlers.ActivateRental)    // Update status
				staff.POST("/rentals/:id/return", handlers.ReturnRental)        // Update status
				staff.POST("/rentals/:id/cancel", handlers.CancelRentalByStaff) // Staff cancel rental
				staff.DELETE("/rentals/:id", handlers.DeleteRental)             // Staff delete rental (use with caution)

				// Payment Management (Staff)
				staff.GET("/payments", handlers.GetPayments)                     // Get all payments
				staff.GET("/rentals/:id/payments", handlers.GetPaymentsByRental) // Get payments for a rental
				staff.POST("/rentals/:id/payments", handlers.ProcessPayment)     // Record payment for a rental

				// Dashboard (Staff)
				staff.GET("/dashboard", handlers.GetDashboard)

				// Note: Review deletion and retrieval by rental ID handled in the general 'protected' group
			}

			// --- Admin Only Routes (Require Admin Role) ---
			adminOnly := protected.Group("/")
			adminOnly.Use(middleware.RoleMiddleware("admin")) // Only admin allowed
			{
				// User Management (Admin - Listing Employees)
				adminOnly.GET("/users", handlers.GetUsers)
				// Potentially add routes for creating/updating/deleting employees here
			}

			// --- Customer Routes (Require Customer Role) ---
			customerOnly := protected.Group("/")
			customerOnly.Use(middleware.CustomerRequired()) // Ensure it's a customer
			{
				// Customer Profile
				customerOnly.GET("/me/profile", handlers.GetMyProfile)
				customerOnly.PUT("/me/profile", handlers.UpdateMyProfile) // Customer updates own profile

				// Customer Rental Actions
				customerOnly.POST("/rentals", handlers.CreateRental)                 // Create new rental booking
				customerOnly.GET("/my/rentals", handlers.GetMyRentals)               // Get own rentals
				customerOnly.GET("/my/rentals/:id", handlers.GetMyRentalByID)        // Get specific own rental
				customerOnly.POST("/my/rentals/:id/cancel", handlers.CancelMyRental) // Customer cancels own rental

				// Customer Review Actions
				customerOnly.POST("/rentals/:id/review", handlers.SubmitReview) // Submit review for own completed rental
				// Note: Review deletion and retrieval by rental ID handled in the general 'protected' group
			}
		}
	}

	// Handle 404 Not Found
	r.NoRoute(func(c *gin.Context) {
		// Respond with JSON for API routes for consistency
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		// Fallback to HTML for non-API routes (if any)
		c.HTML(http.StatusNotFound, "404.html", gin.H{"title": "Page Not Found"})
	})

	log.Println("✅ Routes configured successfully!")
	return r
}

// CORSMiddleware allows cross-origin requests - configure origins properly for production
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// IMPORTANT: Set ALLOWED_ORIGIN environment variable in production
		// to your frontend's actual origin, e.g., "https://your-frontend.com"
		// Using "*" is insecure for production environments.
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			log.Println("⚠️ ALLOWED_ORIGIN environment variable not set. Defaulting to '*' (INSECURE FOR PRODUCTION).")
			allowedOrigin = "*" // Default for local dev, insecure otherwise
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
