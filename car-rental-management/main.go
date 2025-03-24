package main

import (
	"car-rental-management/config"
	"car-rental-management/routes"
)

func main() {
	// Connect to the database
	config.ConnectDB()

	// Set up the routes
	r := routes.SetupRouter()

	r.LoadHTMLGlob("templates/**/*") // ✅ รองรับทุกโฟลเดอร์ย่อย

	// Run the server on port 8080
	r.Run(":8080")
}
