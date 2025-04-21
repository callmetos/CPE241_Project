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
	"github.com/lib/pq" // Import pq for checking specific DB errors
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
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	customer, err := services.GetCustomerByID(id)
	if err != nil {
		if errors.Is(err, errors.New("customer not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("❌ Error fetching customer %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer details"})
		}
		return
	}
	c.JSON(http.StatusOK, customer) // Returns customer without password
}

// UpdateCustomer handles updating customer details (by staff)
func UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var input models.UpdateCustomerByStaffInput // Use the specific input struct for staff
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	updatedCustomer, err := services.UpdateCustomer(id, input) // Pass ID and input struct
	if err != nil {
		log.Println("❌ Error updating customer by staff:", err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to update customer"

		errStr := err.Error()
		if errors.Is(err, errors.New("customer not found for update")) {
			statusCode = http.StatusNotFound
			errMsg = errStr
		} else if errStr == "email already exists for another customer" {
			statusCode = http.StatusConflict
			errMsg = errStr
		} else if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "empty") {
			statusCode = http.StatusBadRequest
			errMsg = errStr
		} else if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // Catch potential wrapped unique constraint error
			statusCode = http.StatusConflict
			errMsg = "Email already exists for another customer"
		} else if strings.Contains(errStr, "update succeeded but failed to fetch") {
			statusCode = http.StatusInternalServerError // Indicate internal issue post-update
			errMsg = "Update may have succeeded, but failed to retrieve final state."
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, updatedCustomer)
}

// DeleteCustomer handles deleting a customer (by staff)
func DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	err = services.DeleteCustomer(id)
	if err != nil {
		log.Println("❌ Error deleting customer by staff:", err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to delete customer"

		errStr := err.Error()
		if errors.Is(err, errors.New("customer not found for deletion")) {
			statusCode = http.StatusNotFound
			errMsg = errStr
		} else if errStr == "cannot delete customer: they have associated rentals" {
			statusCode = http.StatusConflict
			errMsg = errStr
		} else if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23503" { // Catch potential wrapped FK error
			statusCode = http.StatusConflict
			errMsg = "Cannot delete customer: they have associated rentals or other dependencies"
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// --- Customer Self-Service Handlers ---

func GetMyProfile(c *gin.Context) {
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		// This should ideally be caught by CustomerRequired middleware, but double-check
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		log.Printf("🔥 Critical: Invalid customer_id (%v) in context despite passing middleware", customerIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token data"})
		return
	}

	customer, err := services.GetCustomerByID(customerID) // Use existing service function
	if err != nil {
		if errors.Is(err, errors.New("customer not found")) {
			// This should ideally not happen if the JWT is valid and refers to an existing user
			log.Printf("🔥 Critical: Customer ID %d from valid JWT not found in DB!", customerID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile data inconsistent, user may have been deleted"})
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
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		log.Printf("🔥 Critical: Invalid customer_id (%v) in context despite passing middleware", customerIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token data"})
		return
	}

	var input models.UpdateCustomerProfileInput // Use specific input struct for profile update
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Call the specific profile update service function
	updatedCustomer, err := services.UpdateCustomerProfile(customerID, input)
	if err != nil {
		log.Printf("❌ Error updating profile for customer %d: %v", customerID, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to update profile"

		errStr := err.Error()
		// The service returns specific errors for not found or validation issues
		if errors.Is(err, errors.New("customer not found for profile update (or possibly deleted)")) {
			statusCode = http.StatusNotFound // Or maybe Unauthorized if they were deleted?
			errMsg = "Profile not found or user deleted"
		} else if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "empty") {
			statusCode = http.StatusBadRequest
			errMsg = errStr
		} else if strings.Contains(errStr, "profile update succeeded but failed to retrieve") {
			// Update likely worked, but can't confirm full state. Return success but maybe with a warning?
			// Or return 500 as something went wrong post-update. Let's return 500.
			statusCode = http.StatusInternalServerError
			errMsg = "Profile update may have succeeded, but failed to retrieve final state."
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, updatedCustomer)
}
