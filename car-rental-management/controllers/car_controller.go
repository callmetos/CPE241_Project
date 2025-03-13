package controllers

import (
	"car-rental-management/models"
	"car-rental-management/services"
	"net/http"

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
