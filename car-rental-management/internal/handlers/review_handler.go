package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func SubmitReview(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	customerIDInterface, custExists := c.Get("customer_id")

	customerID, ok := customerIDInterface.(int)
	if !custExists || !ok || customerID <= 0 {
		log.Printf("🔥 Invalid customer_id (%v, exists: %v, ok: %v) in context for SubmitReview", customerIDInterface, custExists, ok)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required or invalid token data"})
		return
	}

	var input models.CreateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	review := models.Review{
		RentalID:   rentalID,
		CustomerID: customerID,
		Rating:     input.Rating,
		Comment:    input.Comment,
	}

	createdReview, err := services.CreateReview(review)
	if err != nil {
		log.Printf("❌ Handler: Error submitting review for rental %d by customer %d: %v", rentalID, customerID, err)
		errMsg := err.Error()
		statusCode := http.StatusInternalServerError

		if errors.Is(err, errors.New("rental not found")) {
			statusCode = http.StatusNotFound
			errMsg = "Rental not found"
		} else if errors.Is(err, errors.New("permission denied: you can only review your own rentals")) {
			statusCode = http.StatusForbidden
			errMsg = err.Error()
		} else if errors.Is(err, errors.New("cannot review rental: status is not 'Returned'")) {
			statusCode = http.StatusBadRequest
			errMsg = err.Error()

		} else if errors.Is(err, errors.New("a review for this rental already exists")) || strings.Contains(errMsg, "reviews_rental_id_key") {

			statusCode = http.StatusConflict
			errMsg = "A review for this rental already exists"
		} else if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			statusCode = http.StatusConflict
			errMsg = "A review for this rental already exists"
		} else if strings.Contains(errMsg, "invalid") {
			statusCode = http.StatusBadRequest
		} else {
			errMsg = "Failed to submit review due to an internal error"
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusCreated, createdReview)
}

func GetCarReviews(c *gin.Context) {
	carIDStr := c.Param("id")
	carID, err := strconv.Atoi(carIDStr)
	if err != nil || carID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
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

func GetRentalReview(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	isAllowed := false
	actorRoleInterface, roleExists := c.Get("user_role")
	actorID := 0
	var rental models.Rental

	if !roleExists {
		log.Println("❌ GetRentalReview: user_role not found in context.")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication information missing"})
		return
	}
	actorRole, _ := actorRoleInterface.(string)

	rental, err = services.GetRentalByID(rentalID)
	if err != nil {
		if errors.Is(err, errors.New("rental not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rental not found"})
		} else {
			log.Printf("❌ Error fetching rental %d for review lookup: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rental information"})
		}
		return
	}

	if actorRole == "admin" || actorRole == "manager" {
		isAllowed = true
		empIDInterface, empExists := c.Get("employee_id")
		if empExists {
			if empIDInt, ok := empIDInterface.(int); ok {
				actorID = empIDInt
			}
		}
	} else if actorRole == "customer" {
		custIDInterface, custExists := c.Get("customer_id")
		if custExists {
			if custIDInt, ok := custIDInterface.(int); ok {
				actorID = custIDInt
				if rental.CustomerID == actorID {
					isAllowed = true
				}
			}
		}
	}

	if !isAllowed {
		log.Printf("⚠️ Permission Denied: Actor role '%s' (ID %d) tried to access review for rental %d owned by %d", actorRole, actorID, rentalID, rental.CustomerID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied to view this review"})
		return
	}

	review, err := services.GetReviewByRental(rentalID)
	if err != nil {

		if errors.Is(err, errors.New("review not found for this rental")) || errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Review not found for this rental"})
		} else {
			log.Printf("❌ Handler: Error fetching review for rental %d: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review"})
		}
		return
	}
	c.JSON(http.StatusOK, review)
}

func DeleteReview(c *gin.Context) {
	reviewIDStr := c.Param("id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil || reviewID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var actorID int
	var actorRole string
	var idSet bool

	employeeIDInterface, empExists := c.Get("employee_id")
	userRoleInterface, roleExists := c.Get("user_role")

	if !roleExists {
		log.Println("❌ DeleteReview: user_role not found in context.")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication information missing"})
		return
	}
	roleStr, _ := userRoleInterface.(string)

	if empExists {
		if empIDInt, ok := employeeIDInterface.(int); ok && empIDInt > 0 {
			actorID = empIDInt
			actorRole = roleStr
			idSet = true
		}
	}

	if !idSet {
		customerIDInterface, custExists := c.Get("customer_id")
		if custExists {
			if custIDInt, ok := customerIDInterface.(int); ok && custIDInt > 0 {
				actorID = custIDInt
				actorRole = roleStr
				idSet = true
			}
		}
	}

	if !idSet {

		custIDCheck, custCheckExists := c.Get("customer_id")
		log.Printf("❌ DeleteReview: Could not determine valid actor ID from context. Role: %s, empExists:%v, custExists:%v (value: %v)", roleStr, empExists, custCheckExists, custIDCheck)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token data"})
		return
	}

	err = services.DeleteReview(reviewID, actorID, actorRole)
	if err != nil {
		log.Printf("❌ Handler: Error deleting review %d by actor %d (%s): %v", reviewID, actorID, actorRole, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to delete review"

		errStr := err.Error()
		if errors.Is(err, errors.New("review not found")) {
			statusCode = http.StatusNotFound
			errMsg = errStr
		} else if errors.Is(err, errors.New("permission denied to delete this review")) {
			statusCode = http.StatusForbidden
			errMsg = errStr
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}
