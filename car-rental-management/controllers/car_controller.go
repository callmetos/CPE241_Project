package controllers

import (
	"car-rental-management/models"
	"car-rental-management/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Add a new car with a parking spot
func AddCar(c *gin.Context) {
	var car models.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := services.AddCar(car)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add car"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Car added successfully!"})
}

// Get all available cars
func GetCars(c *gin.Context) {
	cars, err := services.GetAvailableCars()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cars"})
		return
	}
	c.JSON(http.StatusOK, cars)
}

// UpdateCar handles car updates
func UpdateCar(c *gin.Context) {
	var car models.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := services.UpdateCar(car)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update car"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Car updated successfully!"})
}

// DeleteCar handles car deletion
func DeleteCar(c *gin.Context) {
	carIDStr := c.Param("id")            // Gets the ID as a string
	carID, err := strconv.Atoi(carIDStr) // Convert string to int
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	err = services.DeleteCar(carID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete car"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Car deleted successfully!"})
}
