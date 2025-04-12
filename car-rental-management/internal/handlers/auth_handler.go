package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterEmployee handles the registration of a new employee
func RegisterEmployee(c *gin.Context) {
	var employee models.Employee
	if err := c.ShouldBindJSON(&employee); err != nil {
		log.Println("❌ Invalid employee registration data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	log.Printf("📝 Registration attempt for employee: %s", employee.Email)
	err := services.RegisterEmployee(employee)
	if err != nil {
		log.Println("❌ Employee registration failed:", err)
		if err.Error() == "employee email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "password") || strings.Contains(err.Error(), "empty") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register employee"})
		}
		return
	}

	log.Printf("✅ Employee registered successfully: %s", employee.Email)
	c.JSON(http.StatusCreated, gin.H{"message": "Employee registered successfully!"})
}

// LoginEmployee handles the login of an employee
func LoginEmployee(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&credentials); err != nil {
		log.Println("❌ Invalid employee login data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	log.Printf("🔑 Login attempt for employee: %s", credentials.Email)
	token, err := services.AuthenticateEmployee(credentials.Email, credentials.Password)
	if err != nil {
		log.Printf("❌ Employee authentication failed for %s: %v", credentials.Email, err)
		// Return generic error for security
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	log.Printf("✅ Successful login for employee: %s", credentials.Email)
	c.JSON(http.StatusOK, gin.H{"token": token})
}
