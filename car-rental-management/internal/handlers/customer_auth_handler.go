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
	var customer models.Customer

	// Bind JSON request to the customer struct
	if err := c.ShouldBindJSON(&customer); err != nil {
		log.Println("❌ Invalid registration data:", err)
		// Provide more specific binding errors if possible
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration data: " + err.Error()})
		return
	}

	log.Printf("📝 Registration attempt for customer: %s", customer.Email)

	registeredCustomer, err := services.RegisterCustomer(customer)
	if err != nil {
		log.Println("❌ Customer registration failed:", err)
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "empty") || strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "characters long") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register customer"})
		}
		return
	}

	log.Printf("✅ Customer registered successfully: %s", registeredCustomer.Email)
	// Return limited customer info upon registration success (omit password)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful!",
		// Return the created customer object (Password field is empty due to service logic)
		"customer": registeredCustomer,
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	log.Printf("✅ Successful login for customer: %s", credentials.Email)

	// Return the generated JWT token
	c.JSON(http.StatusOK, gin.H{"token": token})
}
