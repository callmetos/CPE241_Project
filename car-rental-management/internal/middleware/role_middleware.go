package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware checks if the authenticated user (assumed Employee) has one of the required roles.
// It should be applied AFTER AuthMiddleware.
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, ensure this is an employee context
		_, empExists := c.Get("employee_id")
		if !empExists {
			log.Println("❌ RoleMiddleware applied, but user is not an employee (no employee_id in context)")
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Employee role required"})
			c.Abort()
			return
		}

		// Get the employee role from context
		userRoleInterface, roleExists := c.Get("user_role")
		role, roleIsString := userRoleInterface.(string)

		if !roleExists || !roleIsString {
			log.Println("❌ No user role found in context or role is not a string (for employee)")
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Role information missing"})
			c.Abort()
			return
		}

		// Check if the role is admin (always allowed for simplicity here)
		// More granular checks might compare allowedRoles even for admin in some cases
		if role == "admin" {
			log.Println("✅ Admin access granted (via RoleMiddleware)")
			c.Next()
			return
		}

		// Check if user role is in the explicitly allowed roles for this route
		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			log.Printf("✅ Access granted for employee role: %s (Allowed: %v)", role, allowedRoles)
			c.Next()
		} else {
			log.Printf("❌ Access denied: Employee role '%s' not in allowed roles %v", role, allowedRoles)
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Insufficient permissions"})
			c.Abort()
		}
	}
}
