package controllers

import (
	"car-rental-management/models"
	"car-rental-management/services"
	"net/http"
	"strconv"

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

// UpdateRental handles rental updates
func UpdateRental(c *gin.Context) {
	var rental models.Rental
	if err := c.ShouldBindJSON(&rental); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := services.UpdateRental(rental)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rental"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rental updated successfully!"})
}

// DeleteRental handles rental deletion
func DeleteRental(c *gin.Context) {
	rentalIDStr := c.Param("id")
	rentalID, err := strconv.Atoi(rentalIDStr) // Convert string to int
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rental ID"})
		return
	}

	err = services.DeleteRental(rentalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rental"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rental deleted successfully!"})
}
