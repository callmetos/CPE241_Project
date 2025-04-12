package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"car-rental-management/internal/utils" // Import utils
	"database/sql"
	"errors"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	// Removed bcrypt import, using utils now
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// --- Employee Auth ---

func RegisterEmployee(employee models.Employee) error {
	// Validation (Role is already checked by model binding oneof)
	if strings.TrimSpace(employee.Name) == "" {
		return errors.New("employee name cannot be empty")
	}
	if !isValidEmail(employee.Email) {
		return errors.New("invalid email format")
	}
	if len(employee.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	// Check existing email
	var count int
	err := config.DB.Get(&count, "SELECT COUNT(*) FROM employees WHERE email=$1", employee.Email)
	if err != nil {
		log.Printf("❌ Database error checking employee email %s: %v", employee.Email, err)
		return errors.New("database error checking email")
	}
	if count > 0 {
		return errors.New("employee email already exists")
	}

	// Hash password using utility function
	hashedPassword, err := utils.HashPassword(employee.Password)
	if err != nil {
		log.Println("❌ Error hashing employee password:", err)
		return errors.New("failed to secure password")
	}
	employee.Password = hashedPassword

	// Insert employee
	query := `INSERT INTO employees (name, email, password, role) VALUES ($1, $2, $3, $4)`
	_, err = config.DB.Exec(query, employee.Name, employee.Email, employee.Password, employee.Role)
	if err != nil {
		log.Println("❌ Error registering employee:", err)
		// Check for specific DB errors like unique constraint violation? (already checked above)
		return errors.New("failed to register employee")
	}

	log.Printf("✅ Employee registered successfully: %s", employee.Email)
	return nil
}

func AuthenticateEmployee(email, password string) (string, error) {
	var employee models.Employee
	query := "SELECT id, name, email, password, role FROM employees WHERE email=$1"
	err := config.DB.QueryRowx(query, email).StructScan(&employee)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("❌ Employee email not found:", email)
		} else {
			log.Println("❌ Error fetching employee:", err)
		}
		return "", errors.New("invalid email or password") // Keep error generic for security
	}

	// Check password using utility function
	if !utils.CheckPasswordHash(password, employee.Password) {
		log.Println("❌ Employee password mismatch for:", email)
		return "", errors.New("invalid email or password") // Keep error generic
	}

	// Generate token (passing employee ID and role)
	token, err := generateEmployeeToken(employee.ID, employee.Email, employee.Role)
	if err != nil {
		log.Println("❌ Error generating employee token:", err)
		return "", errors.New("authentication failed") // Generic error
	}

	log.Printf("✅ Authentication successful for employee: %s", email)
	return token, nil
}

// --- Customer Auth ---

func RegisterCustomer(customer models.Customer) (models.Customer, error) {
	log.Println("🔍 Validating customer data for registration:", customer.Email)
	if strings.TrimSpace(customer.Name) == "" {
		return models.Customer{}, errors.New("customer name cannot be empty")
	}
	if !isValidEmail(customer.Email) {
		return models.Customer{}, errors.New("invalid email format")
	}
	if len(customer.Password) < 6 {
		return models.Customer{}, errors.New("password must be at least 6 characters long")
	}

	var count int
	err := config.DB.Get(&count, "SELECT COUNT(*) FROM customers WHERE email=$1", customer.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) { // Should return 0 if no rows
		log.Println("❌ Error checking existing customer email:", err)
		return models.Customer{}, errors.New("database error checking email existence")
	}
	if count > 0 {
		log.Printf("⚠️ Customer email already exists: %s", customer.Email)
		return models.Customer{}, errors.New("email already exists")
	}

	hashedPassword, err := utils.HashPassword(customer.Password)
	if err != nil {
		log.Println("❌ Error hashing customer password:", err)
		return models.Customer{}, errors.New("failed to secure password")
	}
	customer.Password = hashedPassword

	log.Println("✅ Customer validation passed. Inserting customer...")
	query := `INSERT INTO customers (name, email, phone, password) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	var newID int
	var createdAt, updatedAt time.Time // Assuming time.Time in model
	err = config.DB.QueryRow(query, customer.Name, customer.Email, customer.Phone, customer.Password).Scan(&newID, &createdAt, &updatedAt)
	if err != nil {
		log.Println("❌ Error inserting customer:", err)
		// Check for unique constraint error just in case of race condition?
		return models.Customer{}, errors.New("failed to register customer")
	}

	customer.ID = newID
	customer.CreatedAt = createdAt
	customer.UpdatedAt = updatedAt
	customer.Password = "" // Clear password before returning
	log.Printf("✅ Customer registered successfully with ID: %d", newID)
	return customer, nil
}

func AuthenticateCustomer(email, password string) (string, error) {
	var customer models.Customer
	// Select all fields EXCEPT password hash to avoid accidentally returning it
	query := "SELECT id, name, email, password, phone, created_at, updated_at FROM customers WHERE email=$1"
	err := config.DB.QueryRowx(query, email).StructScan(&customer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("❌ Customer email not found:", email)
		} else {
			log.Println("❌ Error fetching customer:", err)
		}
		return "", errors.New("invalid email or password") // Generic error
	}

	if !utils.CheckPasswordHash(password, customer.Password) {
		log.Println("❌ Customer password mismatch for:", email)
		return "", errors.New("invalid email or password") // Generic error
	}

	token, err := generateCustomerToken(customer.ID, customer.Email)
	if err != nil {
		log.Println("❌ Error generating customer token:", err)
		return "", errors.New("authentication failed")
	}

	log.Printf("✅ Authentication successful for customer: %s", email)
	return token, nil
}

// --- Token Generation ---

func generateEmployeeToken(employeeID int, email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_type":   "employee", // Add user type claim
		"employee_id": employeeID,
		"email":       email,
		"role":        role,
		"exp":         expirationTime.Unix(),
		"iss":         "car-rental-api", // Example issuer
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JwtSecret)) // Use the same secret for now
	if err != nil {
		log.Println("❌ Error signing employee token:", err)
		return "", err
	}
	return tokenString, nil
}

func generateCustomerToken(customerID int, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_type":   "customer", // Add user type claim
		"customer_id": customerID,
		"email":       email,
		"role":        "customer", // Explicit role for consistency
		"exp":         expirationTime.Unix(),
		"iss":         "car-rental-api", // Example issuer
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JwtSecret)) // Use the same secret for now
	if err != nil {
		log.Println("❌ Error signing customer token:", err)
		return "", err
	}
	return tokenString, nil
}
