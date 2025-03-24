package controllers

import (
	"car-rental-management/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDashboard returns rental statistics
func GetDashboard(c *gin.Context) {
	data, err := services.GetDashboardData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard data"})
		return
	}

	c.JSON(http.StatusOK, data)
}
