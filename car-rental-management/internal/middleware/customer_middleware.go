package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CustomerRequired checks if a valid customer_id is set in the context by AuthMiddleware.
// It should be applied AFTER AuthMiddleware on routes intended only for logged-in customers.
func CustomerRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		customerIDInterface, exists := c.Get("customer_id")
		if !exists {
			log.Println("❌ Customer access required, but no customer_id found in context.")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Customer authentication required"})
			c.Abort()
			return
		}
		// Optional: Check if the value is actually an int
		_, ok := customerIDInterface.(int)
		if !ok {
			log.Println("❌ Customer ID in context is not an integer.")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
			c.Abort()
			return
		}

		log.Println("✅ Customer access granted.")
		c.Next()
	}
}
