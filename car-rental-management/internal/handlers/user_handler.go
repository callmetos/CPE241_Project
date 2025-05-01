package handlers // <<< แก้ไขตรงนี้

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	users, err := services.GetUsers()
	if err != nil {
		log.Println("❌ Error fetching users (employees):", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var input models.CreateEmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	createdUser, err := services.CreateEmployeeByAdmin(input)
	if err != nil {
		log.Printf("❌ Handler: Error creating user by admin: %v", err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to create user"

		errStr := err.Error()
		if errStr == "employee email already exists" {
			statusCode = http.StatusConflict
			errMsg = errStr
		} else if errStr == "employee name cannot be empty" || errStr == "invalid email format" || errStr == "password must be at least 6 characters" {
			statusCode = http.StatusBadRequest
			errMsg = errStr
		} else if errStr == "failed to secure password" {

		} else {

			errMsg = errStr
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input models.UpdateEmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	updatedUser, err := services.UpdateEmployeeByAdmin(id, input)
	if err != nil {
		log.Printf("❌ Handler: Error updating user %d by admin: %v", id, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to update user"

		errStr := err.Error()
		if errStr == "employee not found for update" {
			statusCode = http.StatusNotFound
			errMsg = errStr
		} else if errStr == "email already exists for another employee" {
			statusCode = http.StatusConflict
			errMsg = errStr
		} else if errStr == "employee name cannot be empty" || errStr == "invalid email format" {
			statusCode = http.StatusBadRequest
			errMsg = errStr
		} else {
			errMsg = errStr
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = services.DeleteEmployeeByAdmin(id)
	if err != nil {
		log.Printf("❌ Handler: Error deleting user %d by admin: %v", id, err)
		statusCode := http.StatusInternalServerError
		errMsg := "Failed to delete user"

		if errors.Is(err, errors.New("employee not found for deletion")) {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		} else {
			errMsg = err.Error()
		}
		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
