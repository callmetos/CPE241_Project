package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services" // Ensure this import is correct
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPayments(c *gin.Context) {
	payments, err := services.GetPayments()
	if err != nil {
		log.Println("❌ Handler: Error fetching payments:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}
	c.JSON(http.StatusOK, payments)
}

func ProcessPayment(c *gin.Context) {
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

	var input models.RecordPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment data: " + err.Error()})
		return
	}

	createdPayment, err := services.ProcessPayment(rentalID, employeeID, input)
	if err != nil {
		log.Printf("❌ Handler: Error processing payment for rental %d by employee %d: %v", rentalID, employeeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdPayment)
}

func GetPaymentsByRental(c *gin.Context) {
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
			log.Printf("⚠️ GetPaymentsByRental: Error checking rental ownership for rental %d: %v", rentalID, errRent)
			// Use services.ErrRentalNotFound for comparison
			if !errors.Is(errRent, services.ErrRentalNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify rental ownership"})
				return
			}
		}
	}

	if !isAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to view these payments"})
		return
	}

	payments, err := services.GetPaymentsByRentalID(rentalID)
	if err != nil {
		log.Printf("❌ Handler: Error fetching payments for rental %d: %v", rentalID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments for rental"})
		return
	}
	c.JSON(http.StatusOK, payments)
}

func GetPaymentStatus(c *gin.Context) {
	paymentIDStr := c.Param("paymentId")
	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil || paymentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	payment, err := services.GetPaymentStatus(paymentID)
	if err != nil {
		// Assuming GetPaymentStatus in service returns a specific error for "not found"
		if err.Error() == "payment not found" { // Or check with errors.Is if it's a defined error type
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("❌ Handler: Error fetching payment status for ID %d: %v", paymentID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment status"})
		}
		return
	}

	isAllowed := false
	_, empExists := c.Get("employee_id")
	custIDInterface, custExists := c.Get("customer_id")

	if empExists {
		isAllowed = true
	} else if custExists {
		rental, errRent := services.GetRentalByID(payment.RentalID)
		if errRent == nil {
			customerID, ok := custIDInterface.(int)
			if ok && rental.CustomerID == customerID {
				isAllowed = true
			}
		} else {
			log.Printf("⚠️ GetPaymentStatus: Error checking rental ownership for payment %d (rental %d): %v", paymentID, payment.RentalID, errRent)
			if !errors.Is(errRent, services.ErrRentalNotFound) { // Use services.ErrRentalNotFound
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
				return
			}
		}
	}

	if !isAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to view this payment status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"paymentId": payment.ID,
		"rentalId":  payment.RentalID,
		"status":    payment.PaymentStatus,
	})
}

func HandleGetRentalsPendingVerification(c *gin.Context) {
	_, empExists := c.Get("employee_id")
	if !empExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Employee authentication required"})
		return
	}

	rentals, err := services.GetRentalsPendingVerification()
	if err != nil {
		log.Printf("❌ Handler: Error fetching rentals pending verification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rentals pending verification"})
		return
	}
	c.JSON(http.StatusOK, rentals)
}

func HandleVerifyPayment(c *gin.Context) {
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

	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	var input struct {
		Approved bool `json:"approved"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: 'approved' field (boolean) is required"})
		return
	}

	err = services.VerifyPayment(rentalID, input.Approved, employeeID)
	if err != nil {
		log.Printf("❌ Handler: Error verifying payment for rental %d: %v", rentalID, err)
		errMsg := "Failed to verify payment"
		statusCode := http.StatusInternalServerError
		specificErr := err.Error()

		if specificErr == "payment not found or not pending verification" {
			statusCode = http.StatusNotFound
			errMsg = specificErr
		} else if specificErr == "invalid rental or employee ID" { // This error comes from VerifyPayment service
			statusCode = http.StatusBadRequest
			errMsg = specificErr // Use the specific error
		} else if errors.Is(err, services.ErrInvalidState) { // Use services.ErrInvalidState
			statusCode = http.StatusBadRequest
			errMsg = specificErr
		} else {
			errMsg = specificErr
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	action := "approved"
	if !input.Approved {
		action = "rejected"
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Payment for rental %d has been %s.", rentalID, action)})
}
