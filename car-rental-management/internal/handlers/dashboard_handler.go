package handlers

import (
	"car-rental-management/internal/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDashboard returns rental statistics
func GetDashboard(c *gin.Context) {
	data, err := services.GetDashboardData()
	if err != nil {
		log.Println("❌ Error fetching dashboard data:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard data"})
		return
	}
	c.JSON(http.StatusOK, data)
}
