package controllers

import (
	"car-rental-management/models"
	"car-rental-management/services" // Ensure this path is correct
	"net/http"
	"strconv"

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

// UpdateCustomer handles customer updates
func UpdateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := services.UpdateCustomer(customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer updated successfully!"})
}

// DeleteCustomer handles customer deletion
func DeleteCustomer(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, err := strconv.Atoi(customerIDStr) // Convert string to int
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	err = services.DeleteCustomer(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully!"})
}
