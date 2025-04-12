package handlers

import (
	"car-rental-management/internal/services" // <--- ตรวจสอบว่ามี import นี้
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUsers handles listing employees (for admin)
func GetUsers(c *gin.Context) {
	users, err := services.GetUsers() // เรียกใช้ services.GetUsers()
	if err != nil {
		log.Println("❌ Error fetching users (employees):", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}
