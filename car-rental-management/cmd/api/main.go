package main

import (
	// --- ตรวจสอบและแก้ไข Import ให้ถูกต้อง ---
	"car-rental-management/internal/config"
	"car-rental-management/internal/router" // <--- แก้ไขตรงนี้: ลบชื่อเล่น "handlers" ออก

	// --- -------------------------------- ---
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	log.Println("Starting Car Rental API...")

	config.ConnectDB()
	log.Println("Database connection established.")

	// --- ตรวจสอบการเรียกใช้ ---
	r := router.SetupRouter() // <--- เรียกใช้ package router โดยตรง (ถูกต้องแล้ว)
	// --- -------------------- ---
	log.Println("HTTP router initialized.")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Server starting on port %s...", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Could not start server on port %s: %v\n", port, err)
	}

	log.Println("Server stopped.")
}
