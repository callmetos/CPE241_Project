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

// --- Customer Actions ---

// CreateRental handles creating a new rental booking (by customer)
func CreateRental(c *gin.Context) {
	var input models.CreateRentalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Get customer ID from JWT context
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID := customerIDInterface.(int)

	rental := models.Rental{
		CustomerID:      customerID, // Use ID from context
		CarID:           input.CarID,
		PickupDatetime:  input.PickupDatetime,
		DropoffDatetime: input.DropoffDatetime,
		PickupLocation:  input.PickupLocation,
		Status:          "Booked", // Initial status for customer booking
		// BookingDate defaults in DB
	}

	createdRental, err := services.CreateRental(rental)
	if err != nil {
		log.Printf("❌ Error creating rental for customer %d: %v", customerID, err)
		// Handle specific errors from service (car not available, validation, etc.)
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "not available") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rental booking"})
		}
		return
	}

	c.JSON(http.StatusCreated, createdRental) // Return created rental details
}

// GetMyRentals handles listing rentals for the logged-in customer
func GetMyRentals(c *gin.Context) {
	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID := customerIDInterface.(int)

	rentals, err := services.GetRentalsByCustomerID(customerID)
	if err != nil {
		log.Printf("❌ Error fetching rentals for customer %d: %v", customerID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch your rentals"})
		return
	}
	c.JSON(http.StatusOK, rentals)
}

// GetMyRentalByID handles fetching a specific rental for the logged-in customer
func GetMyRentalByID(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID := customerIDInterface.(int)

	rental, err := services.GetRentalByID(rentalID)
	if err != nil {
		if errors.Is(err, errors.New("rental not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("❌ Error fetching rental %d for customer %d: %v", rentalID, customerID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rental details"})
		}
		return
	}

	// Verify ownership
	if rental.CustomerID != customerID {
		log.Printf("⚠️ Permission denied: Customer %d tried to access rental %d owned by %d", customerID, rentalID, rental.CustomerID)
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view this rental"})
		return
	}

	c.JSON(http.StatusOK, rental)
}

// CancelMyRental handles customer cancelling their own rental booking
func CancelMyRental(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	customerIDInterface, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
		return
	}
	customerID := customerIDInterface.(int)

	err = services.CancelCustomerRental(rentalID, customerID)
	if err != nil {
		log.Printf("❌ Error cancelling rental %d by customer %d: %v", rentalID, customerID, err)
		// Handle specific errors from service
		if errors.Is(err, errors.New("rental not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "cannot cancel rental with status") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel rental"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rental cancelled successfully"})
}

// --- Staff Actions ---

// GetRentals handles listing all rentals (for staff)
func GetRentals(c *gin.Context) {
	// TODO: Add filtering by status, customer, car, date range?
	rentals, err := services.GetRentals()
	if err != nil {
		log.Println("❌ Error fetching all rentals:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rentals"})
		return
	}
	c.JSON(http.StatusOK, rentals)
}

// GetRentalByIDForStaff handles fetching any rental by ID (for staff)
func GetRentalByIDForStaff(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}
	rental, err := services.GetRentalByID(rentalID)
	if err != nil {
		if errors.Is(err, errors.New("rental not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("❌ Error fetching rental %d by staff: %v", rentalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rental details"})
		}
		return
	}
	c.JSON(http.StatusOK, rental)
}

func UpdateRentalStatusByStaff(c *gin.Context, targetStatus string) {
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

	// Get Employee ID from context for logging/potential checks in service
	employeeIDInterface, _ := c.Get("employee_id")
	var employeeIDPtr *int
	if employeeIDInt, ok := employeeIDInterface.(int); ok {
		employeeIDPtr = &employeeIDInt
	}

	updatedRental, err := services.UpdateRentalStatus(rentalID, targetStatus, employeeIDPtr)
	if err != nil {
		log.Printf("❌ Handler: Error updating rental %d status to %s by staff: %v", rentalID, targetStatus, err)
		// Handle specific errors from service
		if errors.Is(err, errors.New("rental not found for status update")) || errors.Is(err, errors.New("rental not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rental not found"})
		} else if strings.Contains(err.Error(), "invalid status transition") || strings.Contains(err.Error(), "invalid target status") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "failed to commit") { // Check for commit error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save status update"})
		} else {
			// Default internal server error for other unexpected issues
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rental status"})
		}
		return
	}
	log.Printf("✅ Handler: Rental %d status updated to %s", rentalID, targetStatus)
	c.JSON(http.StatusOK, updatedRental) // Return the updated rental object
}

// --- Specific Status Update Handlers (calling the helper) ---

func ConfirmRental(c *gin.Context) {
	UpdateRentalStatusByStaff(c, "Confirmed")
}

func ActivateRental(c *gin.Context) {
	UpdateRentalStatusByStaff(c, "Active")
}

func ReturnRental(c *gin.Context) {
	UpdateRentalStatusByStaff(c, "Returned")
}

func CancelRentalByStaff(c *gin.Context) {
	UpdateRentalStatusByStaff(c, "Cancelled")
}

// DeleteRental handles deleting a rental (by staff)
func DeleteRental(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}
	err = services.DeleteRental(rentalID)
	if err != nil {
		log.Println("❌ Error deleting rental by staff:", err)
		if errors.Is(err, errors.New("rental not found for deletion")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rental"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rental deleted successfully"})
}
