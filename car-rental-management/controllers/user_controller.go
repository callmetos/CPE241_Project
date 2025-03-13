package controllers

import (
	"car-rental-management/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get all users
func GetUsers(c *gin.Context) {
	users, err := services.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}
