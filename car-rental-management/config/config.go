package config

import (
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB
var JwtSecret string

// ConnectDB sets up the database connection
func ConnectDB() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, using system environment variables")
	}

	// Get the database URL from environment variables
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback to a default connection string for development
		dsn = "postgres://postgres:postgres@localhost:5432/car_rental?sslmode=disable"
		log.Println("⚠️ DATABASE_URL is not set, using default:", dsn)
	}

	// Get the JWT Secret from environment variables
	JwtSecret = os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		// Use a default secret for development (not for production!)
		JwtSecret = "default_development_secret_key_not_for_production"
		log.Println("⚠️ JWT_SECRET is not set, using default secret (not secure for production)")
	}

	log.Println("🔍 Connecting to database...")

	// Try to connect to the PostgreSQL database with retries
	var dbErr error
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		DB, dbErr = sqlx.Connect("postgres", dsn)
		if dbErr == nil {
			break
		}

		log.Printf("❌ Database connection attempt %d failed: %v", i+1, dbErr)
		if i < maxRetries-1 {
			log.Printf("⏳ Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}

	if dbErr != nil {
		log.Fatalf("❌ All database connection attempts failed: %v", dbErr)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := DB.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}

	log.Println("✅ Connected to the database successfully!")

	// Initialize database schema if needed
	initializeSchema()
}

// initializeSchema creates necessary tables if they don't exist
func initializeSchema() {
	log.Println("🔍 Checking database schema...")

	// Create employees table if it doesn't exist
	_, err := DB.Exec(`
	CREATE TABLE IF NOT EXISTS employees (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL
	)`)

	if err != nil {
		log.Fatalf("❌ Failed to create employees table: %v", err)
	}

	// Create cars table if it doesn't exist
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS cars (
		id SERIAL PRIMARY KEY,
		brand VARCHAR(100) NOT NULL,
		model VARCHAR(100) NOT NULL,
		price_per_day DECIMAL(10,2) NOT NULL,
		availability BOOLEAN DEFAULT TRUE,
		parking_spot VARCHAR(50)
	)`)

	if err != nil {
		log.Fatalf("❌ Failed to create cars table: %v", err)
	}

	// Create customers table if it doesn't exist
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		phone VARCHAR(20)
	)`)

	if err != nil {
		log.Fatalf("❌ Failed to create customers table: %v", err)
	}

	// Create rentals table if it doesn't exist
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS rentals (
		id SERIAL PRIMARY KEY,
		customer_id INTEGER REFERENCES customers(id),
		car_id INTEGER REFERENCES cars(id),
		rental_date DATE NOT NULL,
		pickup_datetime TIMESTAMP NOT NULL,
		dropoff_datetime TIMESTAMP NOT NULL,
		pickup_location VARCHAR(200),
		status VARCHAR(50) NOT NULL
	)`)

	if err != nil {
		log.Fatalf("❌ Failed to create rentals table: %v", err)
	}

	// Create payments table if it doesn't exist
	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS payments (
		id SERIAL PRIMARY KEY,
		rental_id INTEGER REFERENCES rentals(id),
		amount DECIMAL(10,2) NOT NULL,
		payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		payment_status VARCHAR(50) NOT NULL
	)`)

	if err != nil {
		log.Fatalf("❌ Failed to create payments table: %v", err)
	}

	log.Println("✅ Database schema initialized successfully!")
}
