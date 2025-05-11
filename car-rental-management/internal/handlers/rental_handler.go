package handlers

import (
	"car-rental-management/internal/models" // Ensure models is imported
	"car-rental-management/internal/services"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetRentals(c *gin.Context) {
	var filters models.RentalFiltersWithPagination // Use models.RentalFiltersWithPagination

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
	filters.SortDirection = c.DefaultQuery("sort_dir", "DESC")

	if rentalIDStr := c.Query("rental_id"); rentalIDStr != "" {
		if rentalID, err := strconv.Atoi(rentalIDStr); err == nil && rentalID > 0 {
			filters.RentalID = &rentalID
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
	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}
	if pickupDateAfterStr := c.Query("pickup_date_after"); pickupDateAfterStr != "" {
		if t, err := time.Parse("2006-01-02", pickupDateAfterStr); err == nil {
			filters.PickupDateAfter = &t
		} else {
			log.Printf("Warning: Invalid pickup_date_after format: %s", pickupDateAfterStr)
			// Optionally return bad request, or ignore filter
		}
	}

	paginatedResponse, err := services.GetRentalsPaginated(filters)
	if err != nil {
		log.Println("Error fetching rentals (paginated):", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rentals"})
		return
	}
	c.JSON(http.StatusOK, paginatedResponse)
}

func GetMyRentals(c *gin.Context) {
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication data"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	rentalsResponse, err := services.GetRentalsByCustomerIDPaginated(customerID, page, limit)
	if err != nil {
		log.Printf("Error fetching rental history for customer %d: %v", customerID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rental history"})
		return
	}
	c.JSON(http.StatusOK, rentalsResponse)
}

func InitiateRental(c *gin.Context) {
	var input models.InitiateRentalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication data"})
		return
	}
	initiatedRental, err := services.InitiateRentalBooking(customerID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to initiate rental booking"
		specificErr := err.Error()
		if errors.Is(err, services.ErrCarNotFound) || errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else if errors.Is(err, services.ErrInvalidDates) || errors.Is(err, services.ErrCarNotAvailable) || errors.Is(err, services.ErrInvalidState) || strings.Contains(specificErr, "overlap") {
			statusCode = http.StatusBadRequest
			errMsg = specificErr
		} else {
			errMsg = specificErr
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Rental initiated successfully. Please proceed to payment.", "id": initiatedRental.ID, "status": initiatedRental.Status})
}

func UploadSlip(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication data"})
		return
	}
	file, err := c.FormFile("slip")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slip file is required: " + err.Error()})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, JPEG, PNG, GIF are allowed."})
		return
	}
	if file.Size > 5*1024*1024 { // 5MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit."})
		return
	}
	uploadDir := filepath.Join(".", "uploads", "slips")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare upload location"})
		return
	}
	uniqueFilename := fmt.Sprintf("rental_%d_cust_%d_%d%s", rentalID, customerID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueFilename)
	fileURL := "/uploads/slips/" + uniqueFilename // This is the URL path client will use
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}
	// Call service to process the slip (e.g., create payment record, update rental status)
	err = services.ProcessSlipUpload(rentalID, customerID, fileURL) // Pass the URL/relative path
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to process slip upload"
		specificErr := err.Error()

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else if errors.Is(err, services.ErrForbidden) {
			statusCode = http.StatusForbidden
			errMsg = specificErr
		} else if errors.Is(err, services.ErrInvalidState) || strings.Contains(specificErr, "cannot upload slip") {
			statusCode = http.StatusBadRequest
			errMsg = specificErr
		} else {
			errMsg = specificErr // Use the specific error message from the service
		}
		// Attempt to remove the saved file if processing failed
		os.Remove(filePath)
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Slip uploaded successfully. Pending verification.", "fileURL": fileURL})
}

func GetRentalByID(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	customerIDInterface, custExists := c.Get("customer_id")
	_, empExists := c.Get("employee_id")

	rental, err := services.GetRentalByID(rentalID)
	if err != nil {
		if errors.Is(err, services.ErrRentalNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Error fetching rental %d: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rental details"})
		}
		return
	}

	isAllowed := false
	if empExists {
		isAllowed = true
	} else if custExists {
		customerID, ok := customerIDInterface.(int)
		if ok && rental.CustomerID == customerID {
			isAllowed = true
		}
	}

	if !isAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view this rental"})
		return
	}
	c.JSON(http.StatusOK, rental)
}

func CancelMyRental(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication data"})
		return
	}

	err = services.CancelCustomerRental(rentalID, customerID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to cancel rental"
		specificErr := err.Error()

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else if errors.Is(err, services.ErrForbidden) {
			statusCode = http.StatusForbidden
			errMsg = specificErr
		} else if errors.Is(err, services.ErrInvalidState) {
			statusCode = http.StatusBadRequest
			errMsg = specificErr
		} else {
			errMsg = specificErr
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rental cancelled successfully"})
}

// UpdateRentalStatusByStaff is a helper function for staff-initiated status changes
func UpdateRentalStatusByStaff(c *gin.Context, targetStatus string) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	employeeIDInterface, exists := c.Get("employee_id")
	if !exists {
		// This should ideally be caught by middleware, but double-check
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Employee authentication required"})
		return
	}
	employeeID, ok := employeeIDInterface.(int)
	if !ok || employeeID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid employee authentication data"})
		return
	}

	log.Printf("🔄 UpdateRentalStatusByStaff: Staff %d attempting to set rental %d status to '%s'", employeeID, rentalID, targetStatus)

	// Pass nil for the tx argument, as the handler doesn't manage it.
	// The service's UpdateRentalStatus will create its own transaction if tx is nil.
	updatedRental, err := services.UpdateRentalStatus(nil, rentalID, targetStatus, &employeeID)
	if err != nil {
		log.Printf("❌ UpdateRentalStatusByStaff: Error updating rental %d to %s: %v", rentalID, targetStatus, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to update rental status"

		specificErr := err.Error()

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else if errors.Is(err, services.ErrInvalidState) || strings.Contains(specificErr, "invalid status transition") {
			statusCode = http.StatusBadRequest
			errMsg = specificErr
		} else if strings.Contains(specificErr, "failed to update car availability") {
			// This specific error from the service indicates a partial success but needs attention
			log.Printf("⚠️ UpdateRentalStatusByStaff: Rental status updated but failed secondary car update for rental %d: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rental status updated, but failed to update car availability. Please check manually.", "rental": updatedRental})
			return
		} else {
			errMsg = specificErr // Use the more specific error from the service layer
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	log.Printf("✅ UpdateRentalStatusByStaff: Rental %d status updated to '%s' successfully", rentalID, targetStatus)
	c.JSON(http.StatusOK, updatedRental)
}

func ConfirmRental(c *gin.Context)       { UpdateRentalStatusByStaff(c, "Confirmed") }
func ActivateRental(c *gin.Context)      { UpdateRentalStatusByStaff(c, "Active") }
func ReturnRental(c *gin.Context)        { UpdateRentalStatusByStaff(c, "Returned") }
func CancelRentalByStaff(c *gin.Context) { UpdateRentalStatusByStaff(c, "Cancelled") }

func DeleteRental(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	// Add logging for staff action
	log.Printf("🗑️ DeleteRental (Staff): Attempting to delete rental %d", rentalID)

	err = services.DeleteRental(rentalID)
	if err != nil {
		log.Printf("❌ DeleteRental (Staff): Error deleting rental %d: %v", rentalID, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to delete rental"
		specificErr := err.Error()

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else {
			// Use the more specific error message from the service layer
			errMsg = specificErr
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	log.Printf("✅ DeleteRental (Staff): Rental %d deleted successfully", rentalID)
	c.JSON(http.StatusOK, gin.H{"message": "Rental deleted successfully"})
}

func GetRentalPrice(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	// Check if user is allowed to see this price
	isAllowed := false
	_, empExists := c.Get("employee_id")
	custIDInterface, custExists := c.Get("customer_id")

	if empExists {
		isAllowed = true // Employees can see any rental price
	} else if custExists {
		// Customers can only see price for their own rentals
		rental, errRent := services.GetRentalByID(rentalID) // Fetch rental to check owner
		if errRent == nil {
			customerID, ok := custIDInterface.(int)
			if ok && rental.CustomerID == customerID {
				isAllowed = true
			}
		} else {
			log.Printf("⚠️ GetRentalPrice: Error checking rental ownership for rental %d: %v", rentalID, errRent)
			if !errors.Is(errRent, services.ErrRentalNotFound) { // Don't expose internal errors, but log them
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify rental ownership"})
				return
			}
			// If rental not found, it will eventually be caught by CalculateRentalCost or the permission check
		}
	}

	if !isAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to view price for this rental"})
		return
	}

	priceDetails, err := services.CalculateRentalCost(rentalID)
	if err != nil {
		log.Printf("❌ GetRentalPrice: Error calculating price for rental %d: %v", rentalID, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to calculate rental price"
		specificErr := err.Error()

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else if errors.Is(err, services.ErrInvalidDates) || strings.Contains(specificErr, "invalid car price") {
			statusCode = http.StatusBadRequest
			errMsg = specificErr
		} else {
			errMsg = specificErr
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rental_id": rentalID, "amount": priceDetails.Amount, "currency": "THB"})
}
