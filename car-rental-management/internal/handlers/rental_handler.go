package handlers

import (
	"car-rental-management/internal/models"
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

func InitiateRental(c *gin.Context) {
	var input models.InitiateRentalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("❌ InitiateRental: Invalid input data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		log.Println("❌ InitiateRental: Customer ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		log.Printf("🔥 InitiateRental: Invalid customer_id type or value in context: %v", customerIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication data"})
		return
	}

	log.Printf("🏁 Initiating rental for customer %d with input: %+v", customerID, input)

	initiatedRental, err := services.InitiateRentalBooking(customerID, input)
	if err != nil {
		log.Printf("❌ InitiateRental: Service error: %v", err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to initiate rental booking"

		errStr := err.Error()
		if errors.Is(err, services.ErrCarNotFound) || errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, services.ErrInvalidDates) || errors.Is(err, services.ErrCarNotAvailable) || errors.Is(err, services.ErrInvalidState) || strings.Contains(errStr, "overlap") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(errStr, "invalid") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{"error": errMsg + ": " + err.Error()})
		return
	}

	log.Printf("✅ InitiateRental: Pending rental created successfully with ID: %d", initiatedRental.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Rental initiated successfully. Please proceed to payment.",
		"id":      initiatedRental.ID,
		"status":  initiatedRental.Status,
	})
}

func UploadSlip(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		log.Printf("❌ UploadSlip: Invalid Rental ID parameter: %s", rentalIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		log.Println("❌ UploadSlip: Customer ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		log.Printf("🔥 UploadSlip: Invalid customer_id type or value in context: %v", customerIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication data"})
		return
	}

	file, err := c.FormFile("slip")
	if err != nil {
		log.Printf("❌ UploadSlip: Error getting form file 'slip': %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slip file is required: " + err.Error()})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[ext] {
		log.Printf("❌ UploadSlip: Invalid file type: %s", ext)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, JPEG, PNG, GIF are allowed."})
		return
	}
	const maxFileSize = 5 * 1024 * 1024
	if file.Size > maxFileSize {
		log.Printf("❌ UploadSlip: File size exceeds limit: %d bytes", file.Size)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit."})
		return
	}

	uploadDir := filepath.Join(".", "uploads", "slips")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("❌ UploadSlip: Failed to create upload directory '%s': %v", uploadDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare upload location"})
		return
	}
	uniqueFilename := fmt.Sprintf("rental_%d_cust_%d_%d%s", rentalID, customerID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueFilename)
	fileURL := "/uploads/slips/" + uniqueFilename

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Printf("❌ UploadSlip: Failed to save uploaded file to '%s': %v", filePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}
	log.Printf("💾 UploadSlip: File saved successfully to: %s", filePath)

	err = services.ProcessSlipUpload(rentalID, customerID, fileURL)
	if err != nil {
		log.Printf("❌ UploadSlip: Service error processing slip upload: %v", err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to process slip upload"

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		} else if errors.Is(err, services.ErrForbidden) {
			statusCode = http.StatusForbidden
			errMsg = err.Error()
		} else if errors.Is(err, services.ErrInvalidState) || strings.Contains(err.Error(), "invalid state") {
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		} else if strings.Contains(err.Error(), "database error") {

		} else {
			errMsg = err.Error()
		}

		removeErr := os.Remove(filePath)
		if removeErr != nil {
			log.Printf("⚠️ UploadSlip: Failed to remove uploaded file '%s' after processing error: %v", filePath, removeErr)
		} else {
			log.Printf("🗑 UploadSlip: Removed uploaded file '%s' due to processing error.", filePath)
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	log.Printf("✅ UploadSlip: Slip uploaded and processed successfully for rental ID: %d", rentalID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Slip uploaded successfully. Pending verification.",
		"fileURL": fileURL,
	})
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
			log.Printf("❌ GetRentalByID: Error fetching rental %d: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rental details"})
		}
		return
	}

	isAllowed := false
	if empExists {
		isAllowed = true
		log.Printf("✅ GetRentalByID: Staff access granted for Rental %d", rentalID)
	} else if custExists {
		customerID, ok := customerIDInterface.(int)
		if ok && rental.CustomerID == customerID {
			isAllowed = true
			log.Printf("✅ GetRentalByID: Customer access granted (CustID: %d) for Rental %d", customerID, rentalID)
		}
	}

	if !isAllowed {
		log.Printf("🚫 GetRentalByID: Access denied for user (CustExists: %t, EmpExists: %t) to Rental %d owned by %d", custExists, empExists, rentalID, rental.CustomerID)
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view this rental"})
		return
	}

	c.JSON(http.StatusOK, rental)
}

func GetMyRentals(c *gin.Context) {
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		log.Println("❌ GetMyRentals: Customer ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}

	customerID, ok := customerIDInterface.(int)
	if !ok || customerID <= 0 {
		log.Printf("🔥 GetMyRentals: Invalid customer_id type or value in context: %v", customerIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication data"})
		return
	}

	log.Printf("🔍 GetMyRentals: Fetching rentals for customer ID: %d", customerID)

	rentals, err := services.GetRentalsByCustomerID(customerID)
	if err != nil {
		log.Printf("❌ GetMyRentals: Error calling GetRentalsByCustomerID for customer %d: %v", customerID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rental history"})
		return
	}

	if rentals == nil {
		rentals = []models.Rental{}
	}

	log.Printf("✅ GetMyRentals: Found %d rentals for customer ID: %d", len(rentals), customerID)
	c.JSON(http.StatusOK, rentals)
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
		log.Printf("❌ CancelMyRental: Error cancelling rental %d for customer %d: %v", rentalID, customerID, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to cancel rental"

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		} else if errors.Is(err, services.ErrForbidden) {
			statusCode = http.StatusForbidden
			errMsg = err.Error()
		} else if strings.Contains(err.Error(), "cannot cancel rental with status") {
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		} else if errors.Is(err, services.ErrInvalidState) {
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	log.Printf("✅ CancelMyRental: Rental %d cancelled successfully by customer %d", rentalID, customerID)
	c.JSON(http.StatusOK, gin.H{"message": "Rental cancelled successfully"})
}

func GetRentals(c *gin.Context) {
	log.Println(" Handler: Fetching all rentals (Staff request)")

	rentals, err := services.GetRentals()
	if err != nil {
		log.Println("❌ GetRentals (Staff): Error fetching rentals:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rentals"})
		return
	}

	if rentals == nil {
		rentals = []models.Rental{}
	}

	log.Printf("✅ GetRentals (Staff): Fetched %d rentals", len(rentals))
	c.JSON(http.StatusOK, rentals)
}

func UpdateRentalStatusByStaff(c *gin.Context, targetStatus string) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	employeeIDInterface, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Employee authentication required"})
		return
	}
	employeeID, ok := employeeIDInterface.(int)
	if !ok || employeeID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid employee authentication data"})
		return
	}

	log.Printf("🔄 UpdateRentalStatusByStaff: Staff %d attempting to set rental %d status to '%s'", employeeID, rentalID, targetStatus)

	updatedRental, err := services.UpdateRentalStatus(rentalID, targetStatus, &employeeID)
	if err != nil {
		log.Printf("❌ UpdateRentalStatusByStaff: Error updating rental %d to %s: %v", rentalID, targetStatus, err)
		statusCode := http.StatusInternalServerError
		var errMsg string

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		} else if errors.Is(err, services.ErrInvalidState) || strings.Contains(err.Error(), "invalid status transition") {
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		} else if strings.Contains(err.Error(), "failed to update car availability") {
			log.Printf("⚠️ UpdateRentalStatusByStaff: Rental status updated but failed secondary car update for rental %d: %v", rentalID, err)

			errMsg = "Rental status updated, but failed to update car availability. Please check manually."
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg, "rental": updatedRental})
			return
		} else {

			errMsg = err.Error()
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

	log.Printf("🗑️ DeleteRental (Staff): Attempting to delete rental %d", rentalID)

	err = services.DeleteRental(rentalID)
	if err != nil {
		log.Printf("❌ DeleteRental (Staff): Error deleting rental %d: %v", rentalID, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to delete rental"

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
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

	isAllowed := false
	_, empExists := c.Get("employee_id")
	custIDInterface, custExists := c.Get("customer_id")

	if empExists {
		isAllowed = true
	} else if custExists {
		rental, errRent := services.GetRentalByID(rentalID)
		if errRent == nil {
			customerID, ok := custIDInterface.(int)
			if ok && rental.CustomerID == customerID {
				isAllowed = true
			}
		} else {
			log.Printf("⚠️ GetRentalPrice: Error checking rental ownership for rental %d: %v", rentalID, errRent)
			if !errors.Is(errRent, services.ErrRentalNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify rental ownership"})
				return
			}
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

		if errors.Is(err, services.ErrRentalNotFound) {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		} else if errors.Is(err, services.ErrInvalidDates) || strings.Contains(err.Error(), "invalid car price") {
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rental_id": rentalID,
		"amount":    priceDetails.Amount,
		"currency":  "THB",
	})
}
