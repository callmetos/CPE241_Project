package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// --- Staff Handlers ---

// GetCustomers handles listing all customers (for staff)
func GetCustomers(c *gin.Context) {
	// TODO: Add pagination? e.g., c.Query("page"), c.Query("limit")
	customers, err := services.GetCustomers()
	if err != nil {
		log.Println("❌ Error fetching customers:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}
	c.JSON(http.StatusOK, customers)
}

// GetCustomerByID handles fetching a single customer (for staff)
func GetCustomerByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	customer, err := services.GetCustomerByID(id)
	if err != nil {
		if errors.Is(err, errors.New("customer not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("❌ Error fetching customer %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer"})
		}
		return
	}
	c.JSON(http.StatusOK, customer) // Returns customer without password
}

// UpdateCustomer handles updating customer details (by staff)
func UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	var customer models.Customer // Staff can update more fields than customer self-update
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	customer.ID = id // Set ID from URL

	updatedCustomer, err := services.UpdateCustomer(customer)
	if err != nil {
		log.Println("❌ Error updating customer by staff:", err)
		if errors.Is(err, errors.New("customer not found for update")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "email") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // Includes duplicate email error
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		}
		return
	}
	c.JSON(http.StatusOK, updatedCustomer)
}

// DeleteCustomer handles deleting a customer (by staff)
func DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	err = services.DeleteCustomer(id)
	if err != nil {
		log.Println("❌ Error deleting customer by staff:", err)
		if errors.Is(err, errors.New("customer not found for deletion")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "cannot delete customer") { // FK error from service
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// --- Customer Self-Service Handlers ---

func GetMyProfile(c *gin.Context) {
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID := customerIDInterface.(int)

	customer, err := services.GetCustomerByID(customerID) // Use existing service function
	if err != nil {
		if errors.Is(err, errors.New("customer not found")) {
			// This should ideally not happen if the JWT is valid and refers to an existing user
			log.Printf("🔥 Critical: Customer ID %d from valid JWT not found in DB!", customerID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile data inconsistent"})
		} else {
			log.Printf("❌ Error fetching profile for customer %d: %v", customerID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		}
		return
	}
	c.JSON(http.StatusOK, customer) // Returns customer data (without password hash)
}

func UpdateMyProfile(c *gin.Context) {
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID := customerIDInterface.(int)

	var input models.UpdateCustomerProfileInput // Use specific input struct for profile update
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Call the specific profile update service function
	updatedCustomer, err := services.UpdateCustomerProfile(customerID, input)
	if err != nil {
		log.Printf("❌ Error updating profile for customer %d: %v", customerID, err)
		if errors.Is(err, errors.New("customer not found for profile update")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"}) // Should not happen
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "empty") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		}
		return
	}
	c.JSON(http.StatusOK, updatedCustomer)
}
