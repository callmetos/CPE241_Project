package config

import (
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sqlx.DB
var JwtSecret string

// ConnectDB sets up the database connection and initializes schema
func ConnectDB() {
	// Load .env file (optional, good for development)
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, relying on system environment variables or defaults.")
	}

	// --- Get Database DSN ---
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/car_rental?sslmode=disable" // Default for local dev
		log.Println("⚠️ DATABASE_URL environment variable not set. Using default:", dsn)
	}

	// --- Get JWT Secret ---
	JwtSecret = os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		JwtSecret = "a_very_secret_dev_key_change_in_production_12345!" // Default for local dev
		log.Println("⚠️ JWT_SECRET environment variable not set. Using default secret (INSECURE FOR PRODUCTION!).")
	}

	// --- Connect to Database with Retry ---
	log.Println("🔍 Connecting to database...")
	var dbErr error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		DB, dbErr = sqlx.Connect("postgres", dsn)
		if dbErr == nil {
			break // Success
		}
		log.Printf("❌ Database connection attempt %d failed: %v", i+1, dbErr)
		if i < maxRetries-1 {
			retryWait := time.Duration(2*(i+1)) * time.Second // Exponential backoff (2s, 4s, 6s...)
			log.Printf("⏳ Retrying connection in %v...", retryWait)
			time.Sleep(retryWait)
		}
	}
	if dbErr != nil {
		log.Fatalf("❌ Failed to connect to database after %d attempts: %v", maxRetries, dbErr)
	}

	// --- Configure Connection Pool ---
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(1 * time.Minute)

	// --- Ping Database ---
	if err := DB.Ping(); err != nil {
		DB.Close()
		log.Fatalf("❌ Cannot ping database: %v", err)
	}

	log.Println("✅ Database connection successful.")

	// --- Initialize Schema (Development Only!) ---
	// IMPORTANT: Comment out or remove this call in production and use migrations.
	runSchemaInitialization := os.Getenv("RUN_SCHEMA_INIT") // Optional flag to control this
	if runSchemaInitialization == "true" {
		initializeSchema()
	} else {
		log.Println("ℹ️ Skipping automatic schema initialization (set RUN_SCHEMA_INIT=true to enable). Use migrations for production.")
	}
}


// initializeSchema creates necessary tables if they don't exist
// WARNING: Development only. Use a proper migration tool (e.g., golang-migrate) for production environments.
func initializeSchema() {
	log.Println("⚠️ Running automatic schema initialization (Development ONLY)...")

	schemaSQL := `
	-- Function to automatically update 'updated_at' timestamp
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
	   NEW.updated_at = NOW(); -- Use NOW() for TIMESTAMPTZ
	   RETURN NEW;
	END;
	$$ language 'plpgsql';

	-- branches table
	CREATE TABLE IF NOT EXISTS branches (
		id SERIAL PRIMARY KEY,
		name VARCHAR(150) NOT NULL UNIQUE,
		address TEXT,
		phone VARCHAR(20),
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	DROP TRIGGER IF EXISTS update_branches_updated_at ON branches;
	CREATE TRIGGER update_branches_updated_at BEFORE UPDATE ON branches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	-- employees table
	CREATE TABLE IF NOT EXISTS employees (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'manager')),
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
	DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;
	CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	-- customers table
	CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		phone VARCHAR(20),
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
	DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;
	CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	-- cars table
	CREATE TABLE IF NOT EXISTS cars (
		id SERIAL PRIMARY KEY,
		brand VARCHAR(100) NOT NULL,
		model VARCHAR(100) NOT NULL,
		price_per_day DECIMAL(10,2) NOT NULL CHECK (price_per_day > 0),
		availability BOOLEAN DEFAULT TRUE,
		parking_spot VARCHAR(50),
		branch_id INT NOT NULL,
		image_url TEXT,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE RESTRICT -- Cannot delete branch if cars exist
	);
	CREATE INDEX IF NOT EXISTS idx_cars_branch_id ON cars(branch_id);
	CREATE INDEX IF NOT EXISTS idx_cars_availability ON cars(availability);
	DROP TRIGGER IF EXISTS update_cars_updated_at ON cars;
	CREATE TRIGGER update_cars_updated_at BEFORE UPDATE ON cars FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	-- rentals table
	CREATE TABLE IF NOT EXISTS rentals (
		id SERIAL PRIMARY KEY,
		customer_id INT NOT NULL,
		car_id INT NOT NULL,
		booking_date DATE DEFAULT CURRENT_DATE,
		pickup_datetime TIMESTAMPTZ NOT NULL,
		dropoff_datetime TIMESTAMPTZ NOT NULL,
		pickup_location TEXT,
		status VARCHAR(50) NOT NULL CHECK (status IN ('Booked', 'Confirmed', 'Active', 'Returned', 'Cancelled')),
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE RESTRICT, -- Cannot delete customer with rentals
		FOREIGN KEY (car_id) REFERENCES cars(id) ON DELETE RESTRICT, -- Cannot delete car with rentals
		CONSTRAINT check_rental_dates CHECK (pickup_datetime < dropoff_datetime)
	);
	CREATE INDEX IF NOT EXISTS idx_rentals_customer_id ON rentals(customer_id);
	CREATE INDEX IF NOT EXISTS idx_rentals_car_id ON rentals(car_id);
	CREATE INDEX IF NOT EXISTS idx_rentals_status ON rentals(status);
	DROP TRIGGER IF EXISTS update_rentals_updated_at ON rentals;
	CREATE TRIGGER update_rentals_updated_at BEFORE UPDATE ON rentals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	-- payments table
	CREATE TABLE IF NOT EXISTS payments (
		id SERIAL PRIMARY KEY,
		rental_id INT NOT NULL,
		amount DECIMAL(10,2) NOT NULL CHECK (amount >= 0),
		payment_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		payment_status VARCHAR(50) NOT NULL CHECK (payment_status IN ('Pending', 'Paid', 'Failed', 'Refunded')),
		payment_method VARCHAR(50),
		recorded_by_employee_id INT,
		transaction_id VARCHAR(100),
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (rental_id) REFERENCES rentals(id) ON DELETE CASCADE, -- Delete payments if rental is deleted
		FOREIGN KEY (recorded_by_employee_id) REFERENCES employees(id) ON DELETE SET NULL -- Keep payment if employee deleted
	);
	CREATE INDEX IF NOT EXISTS idx_payments_rental_id ON payments(rental_id);
	DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
	CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	-- reviews table
	CREATE TABLE IF NOT EXISTS reviews (
		id SERIAL PRIMARY KEY,
		customer_id INT NOT NULL,
		rental_id INT NOT NULL UNIQUE, -- One review per rental
		rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
		comment TEXT,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE, -- Delete review if customer deleted
		FOREIGN KEY (rental_id) REFERENCES rentals(id) ON DELETE CASCADE -- Delete review if rental deleted
	);
	CREATE INDEX IF NOT EXISTS idx_reviews_customer_id ON reviews(customer_id);
	CREATE INDEX IF NOT EXISTS idx_reviews_rental_id ON reviews(rental_id);
	DROP TRIGGER IF EXISTS update_reviews_updated_at ON reviews;
	CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	`

	// Execute the schema SQL
	_, err := DB.Exec(schemaSQL)
	if err != nil {
		log.Fatalf("❌ Failed to execute schema initialization SQL: %v", err)
	}

	log.Println("✅ Database schema initialized successfully (or already exists).")
}