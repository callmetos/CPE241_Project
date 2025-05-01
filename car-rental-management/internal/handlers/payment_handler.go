package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"errors" // Import errors package
	"fmt"    // Import fmt for string formatting
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPayments(c *gin.Context) {
	// Implement logic to fetch all payments (likely admin/manager only)
	payments, err := services.GetPayments()
	if err != nil {
		log.Println("❌ Handler: Error fetching payments:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}
	c.JSON(http.StatusOK, payments)
}

func ProcessPayment(c *gin.Context) { // Handler for Staff recording payment
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
		// Handle specific errors from service if needed
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

	// Permission Check: Allow staff or the customer who owns the rental
	isAllowed := false
	_, empExists := c.Get("employee_id")
	custIDInterface, custExists := c.Get("customer_id")

	if empExists {
		isAllowed = true // Staff can view payments for any rental
	} else if custExists {
		// Check if the customer owns this rental
		rental, errRent := services.GetRentalByID(rentalID) // Use the existing service function
		if errRent == nil {
			customerID, ok := custIDInterface.(int)
			if ok && rental.CustomerID == customerID {
				isAllowed = true
			}
		} else {
			log.Printf("⚠️ GetPaymentsByRental: Error checking rental ownership for rental %d: %v", rentalID, errRent)
			// Proceed with forbidden access if rental check fails, unless it's NotFound
			if !errors.Is(errRent, services.ErrRentalNotFound) { // Use custom error from service
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify rental ownership"})
				return
			}
		}
	}

	if !isAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to view these payments"})
		return
	}

	// Fetch payments if allowed
	payments, err := services.GetPaymentsByRentalID(rentalID)
	if err != nil {
		log.Printf("❌ Handler: Error fetching payments for rental %d: %v", rentalID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments for rental"})
		return
	}
	c.JSON(http.StatusOK, payments)
}

// --- *** Handler สำหรับ GetPaymentStatus *** ---
func GetPaymentStatus(c *gin.Context) {
	paymentIDStr := c.Param("paymentId") // ชื่อ parameter ต้องตรงกับใน router.go
	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil || paymentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	payment, err := services.GetPaymentStatus(paymentID) // Call the service function
	if err != nil {
		// Use the specific error string returned by the service
		if err.Error() == "payment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("❌ Handler: Error fetching payment status for ID %d: %v", paymentID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment status"})
		}
		return
	}

	// --- Permission Check Logic (Example: Staff or Owner) ---
	isAllowed := false
	_, empExists := c.Get("employee_id")
	custIDInterface, custExists := c.Get("customer_id")

	if empExists {
		isAllowed = true // Staff allowed
	} else if custExists {
		// Check if customer owns the rental associated with this payment
		rental, errRent := services.GetRentalByID(payment.RentalID) // Need RentalID from payment object
		if errRent == nil {
			customerID, ok := custIDInterface.(int)
			if ok && rental.CustomerID == customerID {
				isAllowed = true
			}
		} else {
			log.Printf("⚠️ GetPaymentStatus: Error checking rental ownership for payment %d (rental %d): %v", paymentID, payment.RentalID, errRent)
			// Handle error appropriately, maybe deny access if ownership can't be verified
			// For now, we deny if rental check fails (unless NotFound)
			if !errors.Is(errRent, services.ErrRentalNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ownership"})
				return
			}
		}
	}

	if !isAllowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to view this payment status"})
		return
	}
	// --- End Permission Check ---

	// คืนค่าเฉพาะ status หรือข้อมูลที่จำเป็น
	c.JSON(http.StatusOK, gin.H{
		"paymentId": payment.ID,
		"rentalId":  payment.RentalID,
		"status":    payment.PaymentStatus,
	})
}

// --- เพิ่ม: Handler ดึงรายการรอตรวจสอบ ---
func HandleGetRentalsPendingVerification(c *gin.Context) {
	// Middleware RoleMiddleware ควรจะป้องกันไว้แล้ว แต่เช็คอีกครั้งก็ได้
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

// --- เพิ่ม: Handler Approve/Reject ---
func HandleVerifyPayment(c *gin.Context) {
	// ดึง employee ID จาก context
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

	// ดึง rental ID จาก path parameter
	rentalIDStr := c.Param("id") // ต้องตรงกับใน router
	rentalID, err := strconv.Atoi(rentalIDStr)
	if err != nil || rentalID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	// ดึงข้อมูลจาก request body
	var input struct {
		Approved bool `json:"approved"` // คาดหวัง field ชื่อ approved เป็น boolean
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: 'approved' field (boolean) is required"})
		return
	}

	// เรียก Service
	err = services.VerifyPayment(rentalID, input.Approved, employeeID)
	if err != nil {
		log.Printf("❌ Handler: Error verifying payment for rental %d: %v", rentalID, err)
		// แปลง Error จาก Service เป็น HTTP Status ที่เหมาะสม
		errMsg := "Failed to verify payment"
		statusCode := http.StatusInternalServerError
		// Example: check for specific errors from service
		if err.Error() == "payment not found or not pending verification" {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		} else if err.Error() == "invalid rental or employee ID" {
			statusCode = http.StatusBadRequest // หรือ Unauthorized ขึ้นอยู่กับกรณี
		} else if errors.Is(err, services.ErrInvalidState) { // Check for custom error type
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		}
		// เพิ่มเติม Error Handling อื่นๆ
		c.JSON(statusCode, gin.H{"error": errMsg}) // ส่ง Error message จาก service โดยตรง
		return
	}

	action := "approved"
	if !input.Approved {
		action = "rejected"
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Payment for rental %d has been %s.", rentalID, action)})
}
