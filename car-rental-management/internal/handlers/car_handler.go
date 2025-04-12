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

// AddCar handles adding a new car
func AddCar(c *gin.Context) {
	var car models.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Default availability might be better handled in DB or service if needed
	// car.Availability = true

	createdCar, err := services.AddCar(car)
	if err != nil {
		log.Println("❌ Error adding car:", err)
		// Handle specific errors like "branch not found" from service validation
		if strings.Contains(err.Error(), "branch with ID") && strings.Contains(err.Error(), "does not exist") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "empty") || strings.Contains(err.Error(), "greater than zero") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add car"})
		}
		return
	}
	c.JSON(http.StatusCreated, createdCar)
}

// GetCars handles listing cars (potentially filtered)
func GetCars(c *gin.Context) {
	// Example of reading filters from query parameters
	var filters services.CarFilters
	if brand := c.Query("brand"); brand != "" {
		filters.Brand = &brand
	}
	if model := c.Query("model"); model != "" {
		filters.Model = &model
	}
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		if branchID, err := strconv.Atoi(branchIDStr); err == nil {
			filters.BranchID = &branchID
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch_id filter"})
			return
		}
	}
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filters.MinPrice = &minPrice
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_price filter"})
			return
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filters.MaxPrice = &maxPrice
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_price filter"})
			return
		}
	}
	if availabilityStr := c.Query("availability"); availabilityStr != "" {
		if availability, err := strconv.ParseBool(availabilityStr); err == nil {
			filters.Availability = &availability
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid availability filter (must be true or false)"})
			return
		}
	}

	cars, err := services.GetCars(filters) // Use service that accepts filters
	if err != nil {
		log.Println("❌ Error fetching cars:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cars"})
		return
	}
	c.JSON(http.StatusOK, cars)
}

// GetCarByID handles fetching a single car
func GetCarByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}
	car, err := services.GetCarByID(id)
	if err != nil {
		if errors.Is(err, errors.New("car not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("❌ Error fetching car %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch car"})
		}
		return
	}
	c.JSON(http.StatusOK, car)
}

// UpdateCar handles updating car details
func UpdateCar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}
	var car models.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	car.ID = id // Set ID from URL

	updatedCar, err := services.UpdateCar(car)
	if err != nil {
		log.Println("❌ Error updating car:", err)
		if errors.Is(err, errors.New("car not found for update")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "branch") || strings.Contains(err.Error(), "empty") || strings.Contains(err.Error(), "greater than zero") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update car"})
		}
		return
	}
	c.JSON(http.StatusOK, updatedCar)
}

// DeleteCar handles deleting a car
func DeleteCar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}
	err = services.DeleteCar(id)
	if err != nil {
		log.Println("❌ Error deleting car:", err)
		if errors.Is(err, errors.New("car not found for deletion")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "cannot delete car") { // FK error from service
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete car"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Car deleted successfully"})
}
