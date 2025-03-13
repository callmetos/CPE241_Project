package controllers

import (
	"car-rental-management/models"
	"car-rental-management/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get all payments
func GetPayments(c *gin.Context) {
	payments, err := services.GetPayments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}
	c.JSON(http.StatusOK, payments)
}

// Process a new payment
func ProcessPayment(c *gin.Context) {
	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := services.ProcessPayment(payment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Payment processed successfully!"})
}
