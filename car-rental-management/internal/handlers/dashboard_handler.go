package handlers

import (
	"car-rental-management/internal/services"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetDashboard(c *gin.Context) {
	data, err := services.GetDashboardData()
	if err != nil {
		log.Println("❌ Error fetching dashboard data:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard data"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func HandleGetRevenueReport(c *gin.Context) {

	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, errStart := time.Parse("2006-01-02", startDateStr)
	if errStart != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}
	endDate, errEnd := time.Parse("2006-01-02", endDateStr)
	if errEnd != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date cannot be before start_date"})
		return
	}

	reportData, err := services.GetRevenueReport(startDate, endDate)
	if err != nil {
		log.Printf("❌ Handler: Error generating revenue report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate revenue report"})
		return
	}
	c.JSON(http.StatusOK, reportData)
}

func HandleGetPopularCarsReport(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	reportData, err := services.GetPopularCarsReport(limit)
	if err != nil {
		log.Printf("❌ Handler: Error generating popular cars report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate popular cars report"})
		return
	}
	c.JSON(http.StatusOK, reportData)
}

func HandleGetBranchPerformanceReport(c *gin.Context) {
	reportData, err := services.GetBranchPerformanceReport()
	if err != nil {
		log.Printf("❌ Handler: Error generating branch performance report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate branch performance report"})
		return
	}
	c.JSON(http.StatusOK, reportData)
}
