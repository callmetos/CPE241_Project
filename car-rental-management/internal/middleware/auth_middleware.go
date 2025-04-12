package middleware

import (
	"car-rental-management/internal/config" // Use correct path
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens (Employee or Customer)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("❌ Missing Authorization header")
			// Return JSON error for API consistency
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Println("❌ Invalid token format:", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("❌ Unexpected signing method: %v", token.Header["alg"])
				return nil, jwt.ErrSignatureInvalid // Or a more specific error
			}
			return []byte(config.JwtSecret), nil
		})

		if err != nil {
			log.Println("❌ Token error or invalid token:", err)
			errMsg := "Invalid or expired token"
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				errMsg = "Invalid token signature"
			} else if errors.Is(err, jwt.ErrTokenExpired) {
				errMsg = "Token has expired"
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Println("❌ Token is invalid (but parsing succeeded?)")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// --- Check claims for Employee or Customer ---
		email, emailOk := claims["email"].(string)
		userType, userTypeOk := claims["user_type"].(string) // Expect 'employee' or 'customer'

		if !emailOk || !userTypeOk {
			log.Println("❌ Missing required claims (email/user_type) in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("user_email", email) // Set common claim

		if userType == "employee" {
			role, roleOk := claims["role"].(string)
			employeeIDFloat, employeeIDok := claims["employee_id"].(float64) // JWT numbers are often float64

			if !roleOk || !employeeIDok {
				log.Println("❌ Incomplete employee claims in token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid employee token claims"})
				c.Abort()
				return
			}
			employeeID := int(employeeIDFloat)
			log.Printf("✅ Authenticated Employee: %s (ID: %d) with role: %s", email, employeeID, role)
			c.Set("user_role", role)
			c.Set("employee_id", employeeID)

		} else if userType == "customer" {
			customerIDFloat, customerIDok := claims["customer_id"].(float64)
			role, roleOk := claims["role"].(string) // Expect role="customer"

			if !customerIDok || !roleOk || role != "customer" {
				log.Println("❌ Incomplete or invalid customer claims in token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid customer token claims"})
				c.Abort()
				return
			}
			customerID := int(customerIDFloat)
			log.Printf("✅ Authenticated Customer: %s (ID: %d)", email, customerID)
			c.Set("customer_id", customerID)
			c.Set("user_role", "customer") // Set role for consistency if needed
		} else {
			log.Println("❌ Unknown user_type claim in token:", userType)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unrecognized token type"})
			c.Abort()
			return
		}

		c.Next()
	}
}
