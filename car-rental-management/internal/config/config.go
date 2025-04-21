package config

import (
	"log"
	"os"
	"time"

	// Import pq driver for PostgreSQL error codes check (though not used directly here)
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
		log.Println("ℹ️ No .env file found, relying on system environment variables or defaults.")
	}

	// --- Get Database DSN ---
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Provide a more standard default format if possible
		dsn = "postgres://postgres:postgres@localhost:5432/car_rental?sslmode=disable" // Default for local dev
		log.Println("⚠️ DATABASE_URL environment variable not set. Using default:", dsn)
	}

	// --- Get JWT Secret ---
	JwtSecret = os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		JwtSecret = "a_very_secret_dev_key_change_in_production_12345!" // Default for local dev
		log.Println("🚨 CRITICAL WARNING: JWT_SECRET environment variable not set. Using default secret (INSECURE FOR PRODUCTION!). Change this immediately.")
	} else if len(JwtSecret) < 32 {
		log.Println("🚨 SECURITY WARNING: JWT_SECRET is set but seems short. Ensure it is a strong, long, random secret.")
	}

	// --- Connect to Database with Retry ---
	log.Println("🔍 Connecting to database...")
	var dbErr error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		DB, dbErr = sqlx.Connect("postgres", dsn)
		if dbErr == nil {
			// Verify connection is alive
			pingErr := DB.Ping()
			if pingErr == nil {
				break // Success
			}
			log.Printf("❌ Database connected but ping failed: %v", pingErr)
			dbErr = pingErr // Set dbErr to pingErr to trigger retry/fail
			DB.Close()      // Close the potentially problematic connection
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
	DB.SetMaxOpenConns(25)                 // Adjust based on expected load and DB capacity
	DB.SetMaxIdleConns(10)                 // Should be <= MaxOpenConns
	DB.SetConnMaxLifetime(5 * time.Minute) // Recycle connections periodically
	DB.SetConnMaxIdleTime(1 * time.Minute) // Close idle connections sooner

	log.Println("✅ Database connection successful and pool configured.")

	// --- Initialize Schema (Development Only!) ---
	// ######################################################################
	// # CRITICAL WARNING: DEVELOPMENT ONLY!                              #
	// # Do NOT run initializeSchema() in production. Use a proper        #
	// # database migration tool (e.g., golang-migrate/migrate, GORM      #
	// # AutoMigrate carefully, etc.) to manage your schema changes.      #
	// ######################################################################
	runSchemaInitialization := os.Getenv("RUN_SCHEMA_INIT") // Optional flag to control this
	if runSchemaInitialization == "true" {
		initializeSchema()
	} else {
		log.Println("ℹ️ Skipping automatic schema initialization (set RUN_SCHEMA_INIT=true to enable for dev). Use migrations for staging/production.")
	}
}

// initializeSchema creates necessary tables if they don't exist
// ######################################################################
// # WARNING: Development ONLY.                                       #
// # Use a proper migration tool for production environments.           #
// # This function is NOT idempotent for all changes (e.g., index creation) #
// # and may fail if run multiple times without dropping objects first. #
// ######################################################################
func initializeSchema() {
	log.Println("⚠️ Running automatic schema initialization (Development ONLY)...")

	// Use transaction for schema initialization
	tx, err := DB.Beginx()
	if err != nil {
		log.Fatalf("❌ Failed to begin transaction for schema initialization: %v", err)
	}
	// Defer rollback in case of error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Printf("❌ Rolling back schema initialization due to error: %v", err)
			_ = tx.Rollback()
			log.Fatalf("❌ Schema initialization failed.") // Exit after rollback attempt
		} else {
			err = tx.Commit() // Commit if no error
			if err != nil {
				log.Fatalf("❌ Failed to commit schema initialization transaction: %v", err)
			}
			log.Println("✅ Database schema initialized successfully (or already exists).")
		}
	}()

	// Function first, so tables can use it
	triggerFuncSQL := `
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
	   NEW.updated_at = NOW(); -- Use NOW() for TIMESTAMPTZ
	   RETURN NEW;
	END;
	$$ language 'plpgsql';
	`
	_, err = tx.Exec(triggerFuncSQL)
	if err != nil {
		log.Printf("❌ Failed to create/replace trigger function: %v", err)
		// Let defer handle rollback/exit
		return
	}

	// Define schema creation statements
	schemas := map[string]string{
		"branches": `
			CREATE TABLE IF NOT EXISTS branches (
				id SERIAL PRIMARY KEY,
				name VARCHAR(150) NOT NULL UNIQUE,
				address TEXT,
				phone VARCHAR(30), -- Increased length slightly
				created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
			);
			DROP TRIGGER IF EXISTS update_branches_updated_at ON branches;
			CREATE TRIGGER update_branches_updated_at BEFORE UPDATE ON branches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
			CREATE INDEX IF NOT EXISTS idx_branches_name ON branches(name); -- Index on name for lookups
		`,
		"employees": `
			CREATE TABLE IF NOT EXISTS employees (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100) UNIQUE NOT NULL,
				password VARCHAR(255) NOT NULL, -- Store hash
				role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'manager')),
				created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
			DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;
			CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`,
		"customers": `
			CREATE TABLE IF NOT EXISTS customers (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100) UNIQUE NOT NULL,
				phone VARCHAR(30), -- Increased length slightly
				password VARCHAR(255) NOT NULL, -- Store hash
				created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
			DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;
			CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`,
		"cars": `
			CREATE TABLE IF NOT EXISTS cars (
				id SERIAL PRIMARY KEY,
				brand VARCHAR(100) NOT NULL,
				model VARCHAR(100) NOT NULL,
				price_per_day DECIMAL(10,2) NOT NULL CHECK (price_per_day > 0),
				availability BOOLEAN DEFAULT TRUE NOT NULL, -- Ensure not null
				parking_spot VARCHAR(50),
				branch_id INT NOT NULL,
				image_url TEXT,
				created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE RESTRICT -- Cannot delete branch if cars exist
			);
			CREATE INDEX IF NOT EXISTS idx_cars_branch_id ON cars(branch_id);
			CREATE INDEX IF NOT EXISTS idx_cars_availability ON cars(availability);
            CREATE INDEX IF NOT EXISTS idx_cars_brand_model ON cars(brand, model); -- Index for common searches
			DROP TRIGGER IF EXISTS update_cars_updated_at ON cars;
			CREATE TRIGGER update_cars_updated_at BEFORE UPDATE ON cars FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`,
		"rentals": `
			CREATE TABLE IF NOT EXISTS rentals (
				id SERIAL PRIMARY KEY,
				customer_id INT NOT NULL,
				car_id INT NOT NULL,
				booking_date DATE DEFAULT CURRENT_DATE, -- Date only is fine here
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
            CREATE INDEX IF NOT EXISTS idx_rentals_pickup_datetime ON rentals(pickup_datetime); -- Index date ranges
            CREATE INDEX IF NOT EXISTS idx_rentals_dropoff_datetime ON rentals(dropoff_datetime);
			DROP TRIGGER IF EXISTS update_rentals_updated_at ON rentals;
			CREATE TRIGGER update_rentals_updated_at BEFORE UPDATE ON rentals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`,
		"payments": `
			CREATE TABLE IF NOT EXISTS payments (
				id SERIAL PRIMARY KEY,
				rental_id INT NOT NULL,
				amount DECIMAL(10,2) NOT NULL CHECK (amount >= 0),
				payment_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				payment_status VARCHAR(50) NOT NULL CHECK (payment_status IN ('Pending', 'Paid', 'Failed', 'Refunded')),
				payment_method VARCHAR(50),
				recorded_by_employee_id INT, -- Nullable if customer pays online later?
				transaction_id VARCHAR(100) UNIQUE, -- Ensure gateway IDs are unique if present
				created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (rental_id) REFERENCES rentals(id) ON DELETE CASCADE, -- Delete payments if rental is deleted
				FOREIGN KEY (recorded_by_employee_id) REFERENCES employees(id) ON DELETE SET NULL -- Keep payment if employee deleted
			);
			CREATE INDEX IF NOT EXISTS idx_payments_rental_id ON payments(rental_id);
            CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(payment_status);
			DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
			CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`,
		"reviews": `
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
			CREATE INDEX IF NOT EXISTS idx_reviews_rental_id ON reviews(rental_id); -- Already unique, but index helps lookups
			DROP TRIGGER IF EXISTS update_reviews_updated_at ON reviews;
			CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`,
	}

	// Execute schemas in order (though most are independent here)
	// Order matters if there are FK dependencies during creation itself (not the case here)
	tableOrder := []string{"branches", "employees", "customers", "cars", "rentals", "payments", "reviews"}

	for _, tableName := range tableOrder {
		sqlStmt := schemas[tableName]
		log.Printf("Executing schema for: %s...", tableName)
		_, err = tx.Exec(sqlStmt)
		if err != nil {
			log.Printf("❌ Failed to execute schema SQL for table %s: %v", tableName, err)
			// Let defer handle rollback and exit
			return
		}
		log.Printf("✅ Schema applied for: %s", tableName)
	}

	// If loop completes without error, defer will commit.
}
