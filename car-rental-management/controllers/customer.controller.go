package controllers

import (
	"car-rental-management/services" // Ensure this path is correct
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCustomers retrieves all customers from the database
func GetCustomers(c *gin.Context) {
	customers, err := services.GetCustomers() // Ensure this function exists in services package
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}
	c.JSON(http.StatusOK, customers)
}
