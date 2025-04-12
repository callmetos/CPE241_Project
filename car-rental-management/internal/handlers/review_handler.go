package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings" // Needed for error checking

	"github.com/gin-gonic/gin"
)

// SubmitReview handles submitting a review for a specific rental (by customer)
func SubmitReview(c *gin.Context) {
	rentalIDStr := c.Param("id") // Use "id" consistently based on router fix
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID format"})
		return
	}
	if rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID value"})
		return
	}

	customerIDInterface, exists := c.Get("customer_id") // Get customer ID from AuthMiddleware context
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID := customerIDInterface.(int)

	var input models.CreateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	review := models.Review{
		RentalID:   rentalID,
		CustomerID: customerID, // Set from authenticated user
		Rating:     input.Rating,
		Comment:    input.Comment,
	}

	createdReview, err := services.CreateReview(review)
	if err != nil {
		log.Printf("❌ Handler: Error submitting review for rental %d by customer %d: %v", rentalID, customerID, err)
		// --- Improved Error Handling ---
		errMsg := err.Error()
		statusCode := http.StatusInternalServerError // Default to 500

		if errors.Is(err, errors.New("rental not found")) {
			statusCode = http.StatusNotFound
		} else if errMsg == "permission denied: you can only review your own rentals" {
			statusCode = http.StatusForbidden
		} else if errMsg == "cannot review rental: status is not 'Returned'" {
			statusCode = http.StatusBadRequest // Bad request because prerequisite not met
		} else if errMsg == "a review for this rental already exists" {
			statusCode = http.StatusConflict // 409 Conflict
		} else if strings.Contains(errMsg, "invalid") { // Catch other potential validation errors
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		// --- End Improved Error Handling ---
		return
	}
	c.JSON(http.StatusCreated, createdReview)
}

// GetCarReviews handles fetching reviews for a specific car (public)
func GetCarReviews(c *gin.Context) {
	carIDStr := c.Param("id") // Use "id" consistently based on router fix
	carID, err := strconv.Atoi(carIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID format"})
		return
	}
	if carID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID value"})
		return
	}

	reviews, err := services.GetReviewsByCar(carID)
	if err != nil {
		log.Printf("❌ Handler: Error fetching reviews for car %d: %v", carID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}
	c.JSON(http.StatusOK, reviews)
}

// GetRentalReview handles fetching the review for a specific rental (public?)
func GetRentalReview(c *gin.Context) {
	rentalIDStr := c.Param("id") // Use "id" consistently
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID format"})
		return
	}
	if rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID value"})
		return
	}

	// Permission check might be needed here depending on requirements

	review, err := services.GetReviewByRental(rentalID)
	if err != nil {
		log.Printf("❌ Handler: Error fetching review for rental %d: %v", rentalID, err)
		if errors.Is(err, errors.New("review not found for this rental")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review"})
		}
		return
	}
	c.JSON(http.StatusOK, review)
}

// DeleteReview handles deleting a review (customer own, or staff)
func DeleteReview(c *gin.Context) {
	reviewIDStr := c.Param("id") // Assuming route like /reviews/:id
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID format"})
		return
	}
	if reviewID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID value"})
		return
	}

	// Get user info from context to check permissions
	var actorID int
	var actorRole string
	var idSet bool

	customerIDInterface, custExists := c.Get("customer_id")
	employeeIDInterface, empExists := c.Get("employee_id")
	userRoleInterface, roleExists := c.Get("user_role")

	if custExists {
		actorID = customerIDInterface.(int)
		actorRole = "customer"
		idSet = true
	} else if empExists && roleExists {
		actorID = employeeIDInterface.(int)
		actorRole = userRoleInterface.(string)
		idSet = true
	}

	if !idSet {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	err = services.DeleteReview(reviewID, actorID, actorRole)
	if err != nil {
		log.Printf("❌ Handler: Error deleting review %d by actor %d (%s): %v", reviewID, actorID, actorRole, err)
		// Handle permission errors, not found errors from service
		if errors.Is(err, errors.New("review not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if errors.Is(err, errors.New("permission denied to delete this review")) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}
