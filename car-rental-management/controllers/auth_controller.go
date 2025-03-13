package controllers

import (
	"car-rental-management/models"
	"car-rental-management/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterEmployee handles the registration of a new employee
func RegisterEmployee(c *gin.Context) {
	var employee models.Employee

	// Bind the incoming JSON request to the employee struct
	if err := c.ShouldBindJSON(&employee); err != nil {
		log.Println("❌ Invalid request data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Log the registration attempt
	log.Printf("📝 Registration attempt for: %s (%s) with role %s", employee.Name, employee.Email, employee.Role)

	// Set a default role if none provided
	if employee.Role == "" {
		employee.Role = "customer"
		log.Println("⚠️ No role specified, defaulting to 'customer'")
	}

	// Register the employee using the service
	err := services.RegisterEmployee(employee)
	if err != nil {
		log.Println("❌ Registration failed:", err)
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register: " + err.Error()})
		}
		return
	}

	log.Printf("✅ Successfully registered: %s (%s)", employee.Name, employee.Email)
	c.JSON(http.StatusCreated, gin.H{"message": "Employee registered successfully!"})
}

// LoginEmployee handles the login of an employee
func LoginEmployee(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Bind the incoming JSON request to the credentials struct
	if err := c.ShouldBindJSON(&credentials); err != nil {
		log.Println("❌ Invalid login data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request - missing email or password"})
		return
	}

	// Log login attempt (without showing the password)
	log.Printf("🔑 Login attempt for: %s", credentials.Email)

	// Authenticate the employee using the service
	token, err := services.AuthenticateEmployee(credentials.Email, credentials.Password)
	if err != nil {
		log.Printf("❌ Authentication failed for %s: %v", credentials.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Log successful login
	log.Printf("✅ Successful login for: %s", credentials.Email)

	// Return the generated JWT token
	c.JSON(http.StatusOK, gin.H{"token": token})
}
