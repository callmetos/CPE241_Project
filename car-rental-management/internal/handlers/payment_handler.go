package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services" // อาจจะไม่จำเป็นแล้วถ้า error handling เปลี่ยนไป
	"log"
	"net/http"
	"strconv"

	// อาจจะไม่จำเป็นแล้วถ้า error handling เปลี่ยนไป
	"github.com/gin-gonic/gin"
)

// GetPayments handles listing all payments (for staff)
func GetPayments(c *gin.Context) {
	// ... (โค้ดเดิม) ...
}

// ProcessPayment handles recording a payment by staff for a rental
func ProcessPayment(c *gin.Context) {
	// rentalIDStr := c.Param("rental_id") // <--- ของเดิม
	rentalIDStr := c.Param("id") // <--- แก้ไขเป็น "id"
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	// ... (โค้ดส่วนที่เหลือในการดึง employeeID, input และเรียก service) ...
	employeeIDInterface, exists := c.Get("employee_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Employee authentication required"})
		return
	}
	employeeID := employeeIDInterface.(int)

	var input models.RecordPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment data: " + err.Error()})
		return
	}

	createdPayment, err := services.ProcessPayment(rentalID, employeeID, input)
	if err != nil {
		log.Printf("❌ Error processing payment for rental %d by employee %d: %v", rentalID, employeeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}
	c.JSON(http.StatusCreated, createdPayment)
}

// GetPaymentsByRental handles fetching payments for a specific rental
func GetPaymentsByRental(c *gin.Context) {
	// rentalIDStr := c.Param("rental_id") // <--- ของเดิม
	rentalIDStr := c.Param("id") // <--- แก้ไขเป็น "id"
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}
	// ... (โค้ดส่วนที่เหลือในการเช็ค permission และเรียก service) ...
	isAllowed := false
	_, empExists := c.Get("employee_id")
	custIDInterface, custExists := c.Get("customer_id")
	if empExists {
		isAllowed = true
	} else if custExists {
		rental, errRent := services.GetRentalByID(rentalID)
		if errRent == nil && rental.CustomerID == custIDInterface.(int) {
			isAllowed = true
		}
	}
	if !isAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to view these payments"})
		return
	}

	payments, err := services.GetPaymentsByRentalID(rentalID)
	if err != nil {
		log.Printf("❌ Error fetching payments for rental %d: %v", rentalID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments for rental"})
		return
	}
	c.JSON(http.StatusOK, payments)
}
