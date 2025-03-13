package controllers

import (
	"car-rental-management/models"
	"car-rental-management/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Create a new rental
func CreateRental(c *gin.Context) {
	var rental models.Rental
	if err := c.ShouldBindJSON(&rental); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := services.CreateRental(rental)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rental"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rental created successfully!"})
}

// Get all rentals
func GetRentals(c *gin.Context) {
	rentals, err := services.GetRentals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rentals"})
		return
	}
	c.JSON(http.StatusOK, rentals)
}
