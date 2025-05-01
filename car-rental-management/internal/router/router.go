package router

import (
	"car-rental-management/internal/handlers"
	"car-rental-management/internal/middleware"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	r := gin.Default()

	r.Use(CORSMiddleware())
	r.Use(middleware.RequestLogger())

	uploadsDir := "./uploads"
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		log.Printf("⚠️ Uploads directory '%s' not found, creating...", uploadsDir)
		if errMkdir := os.MkdirAll(filepath.Join(uploadsDir, "slips"), 0755); errMkdir != nil {
			log.Printf("🔥 Failed to create uploads directory structure: %v", errMkdir)

		}
	}

	r.Static("/uploads", uploadsDir)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{

			auth.POST("/employee/login", handlers.LoginEmployee)
			auth.POST("/customer/register", handlers.RegisterCustomer)
			auth.POST("/customer/login", handlers.LoginCustomer)
		}

		api.GET("/cars", handlers.GetCars)
		api.GET("/cars/:id", handlers.GetCarByID)
		api.GET("/cars/:id/reviews", handlers.GetCarReviews)
		api.GET("/branches", handlers.GetBranches)
		api.GET("/branches/:id", handlers.GetBranchByID)

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/rentals/:id", handlers.GetRentalByID)
			protected.GET("/rentals/:id/price", handlers.GetRentalPrice)
			protected.GET("/payments/:paymentId/status", handlers.GetPaymentStatus)
			protected.DELETE("/reviews/:id", handlers.DeleteReview)
			protected.GET("/rentals/:id/review", handlers.GetRentalReview)

			staff := protected.Group("/")
			staff.Use(middleware.RoleMiddleware("admin", "manager"))
			{
				staff.POST("/branches", handlers.CreateBranch)
				staff.PUT("/branches/:id", handlers.UpdateBranch)
				staff.DELETE("/branches/:id", handlers.DeleteBranch)

				staff.POST("/cars", handlers.AddCar)
				staff.PUT("/cars/:id", handlers.UpdateCar)
				staff.DELETE("/cars/:id", handlers.DeleteCar)

				staff.GET("/customers", handlers.GetCustomers)
				staff.GET("/customers/:id", handlers.GetCustomerByID)
				staff.PUT("/customers/:id", handlers.UpdateCustomer)
				staff.DELETE("/customers/:id", handlers.DeleteCustomer)

				staff.GET("/rentals", handlers.GetRentals)
				staff.POST("/rentals/:id/confirm", handlers.ConfirmRental)
				staff.POST("/rentals/:id/activate", handlers.ActivateRental)
				staff.POST("/rentals/:id/return", handlers.ReturnRental)
				staff.POST("/rentals/:id/cancel", handlers.CancelRentalByStaff)
				staff.DELETE("/rentals/:id", handlers.DeleteRental)

				staff.GET("/payments", handlers.GetPayments)
				staff.GET("/rentals/:id/payments", handlers.GetPaymentsByRental)
				staff.POST("/rentals/:id/payments", handlers.ProcessPayment)

				staff.GET("/rentals/pending-verification", handlers.HandleGetRentalsPendingVerification)
				staff.POST("/rentals/:id/verify-payment", handlers.HandleVerifyPayment)

				staff.GET("/dashboard", handlers.GetDashboard)

				reports := staff.Group("/reports")
				{
					reports.GET("/revenue", handlers.HandleGetRevenueReport)
					reports.GET("/popular-cars", handlers.HandleGetPopularCarsReport)
					reports.GET("/branch-performance", handlers.HandleGetBranchPerformanceReport)
				}

			}

			adminOnly := protected.Group("/")
			adminOnly.Use(middleware.RoleMiddleware("admin"))
			{

				adminOnly.GET("/users", handlers.GetUsers)

				adminOnly.POST("/users", handlers.CreateUser)
				adminOnly.PUT("/users/:id", handlers.UpdateUser)
				adminOnly.DELETE("/users/:id", handlers.DeleteUser)

			}

			customerOnly := protected.Group("/")
			customerOnly.Use(middleware.CustomerRequired())
			{
				customerOnly.GET("/me/profile", handlers.GetMyProfile)
				customerOnly.PUT("/me/profile", handlers.UpdateMyProfile)

				customerOnly.POST("/rentals/initiate", handlers.InitiateRental)
				customerOnly.POST("/rentals/:id/upload-slip", handlers.UploadSlip)

				customerOnly.GET("/my/rentals", handlers.GetMyRentals)

				customerOnly.POST("/my/rentals/:id/cancel", handlers.CancelMyRental)
				customerOnly.POST("/rentals/:id/review", handlers.SubmitReview)
			}
		}
	}

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API resource not found"})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})

	})

	log.Println("✅ Routes configured successfully!")
	return r
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			log.Println("⚠️ ALLOWED_ORIGIN environment variable not set. Defaulting to '*' (INSECURE FOR PRODUCTION). Set it to your frontend URL (e.g., http://localhost:5173).")
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
