package middleware

import (
	"car-rental-management/config"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware is used to validate JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("❌ Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
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
			return []byte(config.JwtSecret), nil
		})

		if err != nil {
			log.Println("❌ Token parsing error:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Println("❌ Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract user information from claims
		email, emailOk := claims["email"].(string)
		role, roleOk := claims["role"].(string)

		if !emailOk || !roleOk {
			log.Println("❌ Missing claims in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		log.Printf("✅ Authenticated user: %s with role: %s", email, role)
		c.Set("user_email", email)
		c.Set("user_role", role)
		c.Next()
	}
}

// RoleMiddleware checks if the user has the required role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			log.Println("❌ No user role found in context")
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: No role assigned"})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			log.Println("❌ User role is not a string")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: Invalid role type"})
			c.Abort()
			return
		}

		log.Printf("🔒 Checking role access: User role %s, Allowed roles %v", role, allowedRoles)

		// Admin role has access to everything
		if role == "admin" {
			log.Println("✅ Admin access granted")
			c.Next()
			return
		}

		// Check if user role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				log.Printf("✅ Access granted for role: %s", role)
				c.Next()
				return
			}
		}

		log.Printf("❌ Access denied: User role %s not in allowed roles %v", role, allowedRoles)
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
		c.Abort()
	}
}
