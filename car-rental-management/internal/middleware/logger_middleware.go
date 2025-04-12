// สามารถสร้างไฟล์ใหม่ชื่อ internal/middleware/logger_middleware.go
// หรือเพิ่มฟังก์ชันนี้เข้าไปในไฟล์ middleware อื่นที่มีอยู่แล้ว เช่น auth_middleware.go

package middleware // <--- ต้องเป็น package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs basic info about incoming requests and their responses
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- Start Timer ---
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// --- Process Request ---
		c.Next() // <--- ให้ Handler อื่นๆ ทำงานต่อไป

		// --- Log Results ---
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		// errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String() // ดึง Error ภายใน (ถ้ามี)

		logLine := fmt.Sprintf("[REQUEST] %s | %3d | %13v | %15s | %-7s %s",
			time.Now().Format("2006/01/02 - 15:04:05"), // Format timestamp
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
		if raw != "" {
			logLine += "?" + raw
		}
		// if errorMessage != "" {
		//	 logLine += " | " + errorMessage // เพิ่ม Error message ใน Log (ถ้าต้องการ)
		// }

		log.Println(logLine) // พิมพ์ Log ออกมา
	}
}
