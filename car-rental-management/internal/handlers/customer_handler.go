package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings" // Import strings

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func GetCustomers(c *gin.Context) {
	var filters services.CustomerFiltersWithPagination

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, errPage := strconv.Atoi(pageStr)
	limit, errLimit := strconv.Atoi(limitStr)

	if errPage != nil || page <= 0 {
		page = 1
	}
	if errLimit != nil || limit <= 0 {
		limit = 10
	}
	filters.Page = page
	filters.Limit = limit

	filters.SortBy = c.DefaultQuery("sort_by", "id")
	filters.SortDirection = c.DefaultQuery("sort_dir", "ASC")

	if name := c.Query("name"); name != "" {
		filters.Name = &name
	}
	if email := c.Query("email"); email != "" {
		filters.Email = &email
	}
	if phone := c.Query("phone"); phone != "" {
		filters.Phone = &phone
	}

	paginatedResponse, err := services.GetCustomersPaginated(filters)
	if err != nil {
		log.Println("Error fetching customers (paginated):", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}
	c.JSON(http.StatusOK, paginatedResponse)
}

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
			log.Printf("Error fetching customer %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer details"})
		}
		return
	}
	c.JSON(http.StatusOK, customer)
}

func UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var input models.UpdateCustomerByStaffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	updatedCustomer, err := services.UpdateCustomer(id, input)
	if err != nil {
		log.Println("Error updating customer by staff:", err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to update customer"

		// Use the err.Error() directly or specific checks
		specificErr := err.Error() // Assign to a variable to avoid re-calling .Error()
		if errors.Is(err, errors.New("customer not found for update")) {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else if specificErr == "email already exists for another customer" {
			statusCode = http.StatusConflict
			errMsg = specificErr
		} else if strings.Contains(specificErr, "invalid") || strings.Contains(specificErr, "empty") {
			statusCode = http.StatusBadRequest
			errMsg = specificErr
		} else if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			statusCode = http.StatusConflict
			errMsg = "Email already exists for another customer"
		} else if strings.Contains(specificErr, "update succeeded but failed to fetch") {
			statusCode = http.StatusInternalServerError
			errMsg = "Update may have succeeded, but failed to retrieve final state."
		} else {
			errMsg = specificErr // Default to the error message from service
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, updatedCustomer)
}

func DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}
	err = services.DeleteCustomer(id)
	if err != nil {
		log.Println("Error deleting customer by staff:", err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to delete customer"

		specificErr := err.Error()
		if errors.Is(err, errors.New("customer not found for deletion")) {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else if specificErr == "cannot delete customer: they have associated rentals" {
			statusCode = http.StatusConflict
			errMsg = specificErr
		} else if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23503" {
			statusCode = http.StatusConflict
			errMsg = "Cannot delete customer: they have associated rentals or other dependencies"
		} else {
			errMsg = specificErr
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

func GetMyProfile(c *gin.Context) {
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		log.Printf("Critical: Invalid customer_id (%v) in context despite passing middleware", customerIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token data"})
		return
	}

	customer, err := services.GetCustomerByID(customerID)
	if err != nil {
		if errors.Is(err, errors.New("customer not found")) {
			log.Printf("Critical: Customer ID %d from valid JWT not found in DB!", customerID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile data inconsistent, user may have been deleted"})
		} else {
			log.Printf("Error fetching profile for customer %d: %v", customerID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		}
		return
	}
	c.JSON(http.StatusOK, customer)
}

func UpdateMyProfile(c *gin.Context) {
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		log.Printf("Critical: Invalid customer_id (%v) in context despite passing middleware", customerIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token data"})
		return
	}

	var input models.UpdateCustomerProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	updatedCustomer, err := services.UpdateCustomerProfile(customerID, input)
	if err != nil {
		log.Printf("Error updating profile for customer %d: %v", customerID, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to update profile"

		specificErr := err.Error()
		if errors.Is(err, errors.New("customer not found for profile update (or possibly deleted)")) {
			statusCode = http.StatusNotFound
			errMsg = "Profile not found or user deleted"
		} else if strings.Contains(specificErr, "invalid") || strings.Contains(specificErr, "empty") {
			statusCode = http.StatusBadRequest
			errMsg = specificErr
		} else if strings.Contains(specificErr, "profile update succeeded but failed to retrieve") {
			statusCode = http.StatusInternalServerError
			errMsg = "Profile update may have succeeded, but failed to retrieve final state."
		} else {
			errMsg = specificErr
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, updatedCustomer)
}
