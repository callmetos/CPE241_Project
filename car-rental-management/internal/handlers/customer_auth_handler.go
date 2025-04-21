package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterCustomer handles customer registration requests
func RegisterCustomer(c *gin.Context) {
	var input models.RegisterCustomerInput // Use the specific input struct

	// Bind JSON request to the input struct
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("❌ Invalid registration data:", err)
		// Provide more specific binding errors if possible
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration data: " + err.Error()})
		return
	}

	log.Printf("📝 Registration attempt for customer: %s", input.Email)

	// Call service with the input struct
	registeredCustomer, err := services.RegisterCustomer(input)
	if err != nil {
		log.Println("❌ Customer registration failed:", err)
		statusCode := http.StatusInternalServerError // Default
		errMsg := "Failed to register customer"

		// Map specific service errors to HTTP statuses
		errStr := err.Error()
		if errStr == "email already exists" {
			statusCode = http.StatusConflict
			errMsg = errStr
		} else if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "empty") || strings.Contains(errStr, "password") || strings.Contains(errStr, "characters long") {
			statusCode = http.StatusBadRequest
			errMsg = errStr
		}
		// Check for wrapped DB errors if needed, though generic 500 might be okay here

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	log.Printf("✅ Customer registered successfully: %s", registeredCustomer.Email)
	// Return limited customer info upon registration success (omit password)
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Registration successful!",
		"customer": registeredCustomer, // Password field is already cleared by the service
	})
}

// LoginCustomer handles customer login requests
func LoginCustomer(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// Bind JSON request to credentials struct
	if err := c.ShouldBindJSON(&credentials); err != nil {
		log.Println("❌ Invalid login data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data: " + err.Error()})
		return
	}

	log.Printf("🔑 Login attempt for customer: %s", credentials.Email)

	// Authenticate customer using the service
	token, err := services.AuthenticateCustomer(credentials.Email, credentials.Password)
	if err != nil {
		log.Printf("❌ Customer authentication failed for %s: %v", credentials.Email, err)
		// Check for specific error types if needed, otherwise return general invalid credentials
		// The service already returns "invalid email or password" for not found or wrong password.
		// Check for potential wrapped DB errors from fetching?
		if strings.Contains(err.Error(), "error fetching customer data") {
			// Internal server error if DB fetch failed
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed due to an internal error"})
		} else {
			// Primarily "invalid email or password" or "authentication failed" (token signing)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}
		return
	}

	log.Printf("✅ Successful login for customer: %s", credentials.Email)

	// Return the generated JWT token
	c.JSON(http.StatusOK, gin.H{"token": token})
}
