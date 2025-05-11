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

func AddCar(c *gin.Context) {
	var car models.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	createdCar, err := services.AddCar(car)
	if err != nil {
		log.Println("Error adding car:", err)
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

func GetCars(c *gin.Context) {
	var filters services.CarFiltersWithPagination

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, errPage := strconv.Atoi(pageStr)
	limit, errLimit := strconv.Atoi(limitStr)

	if errPage != nil || page <= 0 {
		page = 1
	}
	if errLimit != nil || limit <= 0 {
		limit = 10
	}
	filters.Page = page
	filters.Limit = limit

	filters.SortBy = c.DefaultQuery("sort_by", "id")
	filters.SortDirection = c.DefaultQuery("sort_dir", "ASC")

	if brand := c.Query("brand"); brand != "" {
		filters.Brand = &brand
	}
	if model := c.Query("model"); model != "" {
		filters.Model = &model
	}
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" {
		if branchID, err := strconv.Atoi(branchIDStr); err == nil && branchID > 0 {
			filters.BranchID = &branchID
		} else if branchIDStr != "" {
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

	paginatedResponse, err := services.GetCarsPaginated(filters)
	if err != nil {
		log.Println("Error fetching cars (paginated):", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cars"})
		return
	}
	c.JSON(http.StatusOK, paginatedResponse)
}

func GetCarByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}
	car, err := services.GetCarByID(id)
	if err != nil {
		if errors.Is(err, errors.New("car not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Error fetching car %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch car"})
		}
		return
	}
	c.JSON(http.StatusOK, car)
}

func UpdateCar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}
	var car models.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	car.ID = id

	updatedCar, err := services.UpdateCar(car)
	if err != nil {
		log.Println("Error updating car:", err)
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

func DeleteCar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}
	err = services.DeleteCar(id)
	if err != nil {
		log.Println("Error deleting car:", err)
		if errors.Is(err, errors.New("car not found for deletion")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "cannot delete car") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete car"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Car deleted successfully"})
}
