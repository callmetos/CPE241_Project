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

	userRoleInterface, roleExists := c.Get("user_role")
	if !roleExists {
		log.Println("❌ GetRentalReview: user_role not found in context.")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication information missing"})
		return
	}
	actorRole, _ := userRoleInterface.(string)

	isAllowed := false
	var actorIDForLog int // Used for logging in case of permission denial

	// Fetch rental details first to check ownership if the actor is a customer
	rental, err := services.GetRentalByID(rentalID) // Assuming GetRentalByID fetches CustomerID
	if err != nil {
		if errors.Is(err, services.ErrRentalNotFound) { // Use specific error from service
			c.JSON(http.StatusNotFound, gin.H{"error": "Rental not found"})
		} else {
			log.Printf("❌ Error fetching rental %d for review lookup: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rental information"})
		}
		return
	}

	if actorRole == "admin" || actorRole == "manager" {
		isAllowed = true
		// Get employee_id for logging, though not strictly for permission logic itself here
		if empIDInterface, empExists := c.Get("employee_id"); empExists {
			if idVal, ok := empIDInterface.(int); ok {
				actorIDForLog = idVal // idVal is used by assigning to actorIDForLog
			}
		}
	} else if actorRole == "customer" {
		if custIDInterface, custExists := c.Get("customer_id"); custExists {
			if idVal, ok := custIDInterface.(int); ok {
				actorIDForLog = idVal // idVal is used by assigning to actorIDForLog
				if rental.CustomerID == actorIDForLog {
					isAllowed = true
				}
			}
		}
	}

	if !isAllowed {
		// actorIDForLog is used here
		log.Printf("⚠️ Permission Denied: Actor role '%s' (ID %d) tried to access review for rental %d owned by Customer %d",
			actorRole, actorIDForLog, rentalID, rental.CustomerID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied to view this review"})
		return
	}

	// If allowed, proceed to get the review
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
		if idVal, ok := employeeIDInterface.(int); ok && idVal > 0 { // Renamed empIDInt to idVal
			actorID = idVal // idVal is used here
			actorRole = roleStr
			idSet = true
		}
	}

	if !idSet {
		customerIDInterface, custExists := c.Get("customer_id")
		if custExists {
			if idVal, ok := customerIDInterface.(int); ok && idVal > 0 { // Renamed custIDInt to idVal
				actorID = idVal // idVal is used here
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
		} else {
			errMsg = errStr // Use the specific error message
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}

func HandleGetAllReviewsAdmin(c *gin.Context) {
	var filters models.ReviewFiltersWithPagination

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

	filters.SortBy = c.DefaultQuery("sort_by", "review_created_at")
	filters.SortDirection = c.DefaultQuery("sort_dir", "DESC")

	if ratingStr := c.Query("rating"); ratingStr != "" {
		if rating, err := strconv.Atoi(ratingStr); err == nil && rating >= 1 && rating <= 5 {
			filters.Rating = &rating
		}
	}
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if customerID, err := strconv.Atoi(customerIDStr); err == nil && customerID > 0 {
			filters.CustomerID = &customerID
		}
	}
	if carIDStr := c.Query("car_id"); carIDStr != "" {
		if carID, err := strconv.Atoi(carIDStr); err == nil && carID > 0 {
			filters.CarID = &carID
		}
	}
	if keyword := c.Query("keyword"); keyword != "" {
		filters.Keyword = &keyword
	}

	paginatedResponse, err := services.GetAllReviewsPaginated(filters)
	if err != nil {
		log.Println("❌ Error fetching all reviews for admin:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}
	c.JSON(http.StatusOK, paginatedResponse)
}
