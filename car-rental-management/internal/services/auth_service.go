package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"car-rental-management/internal/utils" // Import utils
	"database/sql"
	"errors"
	"fmt" // Import fmt for error wrapping
	"log"

	// Removed regexp import
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- Employee Auth ---

func RegisterEmployee(employee models.Employee) error {
	// Validation (Role is already checked by model binding oneof)
	if strings.TrimSpace(employee.Name) == "" {
		return errors.New("employee name cannot be empty")
	}
	// Use validation util
	if !utils.IsValidEmail(employee.Email) {
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
		// Wrap the error
		return fmt.Errorf("database error checking email: %w", err)
	}
	if count > 0 {
		return errors.New("employee email already exists")
	}

	// Hash password using utility function
	hashedPassword, err := utils.HashPassword(employee.Password)
	if err != nil {
		log.Println("❌ Error hashing employee password:", err)
		// Consider wrapping if HashPassword returns specific error types
		return errors.New("failed to secure password")
	}
	employee.Password = hashedPassword

	// Insert employee
	query := `INSERT INTO employees (name, email, password, role) VALUES ($1, $2, $3, $4)`
	_, err = config.DB.Exec(query, employee.Name, employee.Email, employee.Password, employee.Role)
	if err != nil {
		log.Println("❌ Error registering employee:", err)
		// Wrap the error
		return fmt.Errorf("failed to register employee: %w", err)
	}

	log.Printf("✅ Employee registered successfully: %s", employee.Email)
	return nil
}

func AuthenticateEmployee(email, password string) (string, error) {
	var employee models.Employee
	query := "SELECT id, name, email, password, role FROM employees WHERE email=$1"
	// Use Get instead of QueryRowx/StructScan for simpler error checking
	err := config.DB.Get(&employee, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ Employee email not found: %s", email)
			return "", errors.New("invalid email or password") // Keep error generic
		}
		log.Printf("❌ Error fetching employee %s: %v", email, err)
		// Wrap the database error
		return "", fmt.Errorf("error fetching employee data: %w", err)
	}

	// Check password using utility function
	if !utils.CheckPasswordHash(password, employee.Password) {
		log.Printf("❌ Employee password mismatch for: %s", email)
		return "", errors.New("invalid email or password") // Keep error generic
	}

	// Generate token (passing employee ID and role)
	token, err := generateEmployeeToken(employee.ID, employee.Email, employee.Role)
	if err != nil {
		log.Println("❌ Error generating employee token:", err)
		// Don't wrap internal token generation error usually, return generic auth failure
		return "", errors.New("authentication failed")
	}

	log.Printf("✅ Authentication successful for employee: %s", email)
	return token, nil
}

// --- Customer Auth ---

// RegisterCustomer now accepts RegisterCustomerInput
func RegisterCustomer(input models.RegisterCustomerInput) (models.Customer, error) {
	log.Println("🔍 Validating customer data for registration:", input.Email)
	// Validation performed via binding tags in the handler mostly
	// Re-check here for service-level assurance if needed, but redundant if binding is robust
	if strings.TrimSpace(input.Name) == "" {
		return models.Customer{}, errors.New("customer name cannot be empty")
	}
	if !utils.IsValidEmail(input.Email) {
		return models.Customer{}, errors.New("invalid email format")
	}
	if len(input.Password) < 6 {
		return models.Customer{}, errors.New("password must be at least 6 characters long")
	}

	// Check existing email
	var count int
	err := config.DB.Get(&count, "SELECT COUNT(*) FROM customers WHERE email=$1", input.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) { // Should return 0 if no rows
		log.Println("❌ Error checking existing customer email:", err)
		return models.Customer{}, fmt.Errorf("database error checking email existence: %w", err)
	}
	if count > 0 {
		log.Printf("⚠️ Customer email already exists: %s", input.Email)
		return models.Customer{}, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		log.Println("❌ Error hashing customer password:", err)
		return models.Customer{}, errors.New("failed to secure password")
	}

	log.Println("✅ Customer validation passed. Inserting customer...")
	// Create the customer model to insert
	customer := models.Customer{
		Name:     input.Name,
		Email:    input.Email,
		Phone:    input.Phone,
		Password: hashedPassword, // Use the hash
	}

	query := `INSERT INTO customers (name, email, phone, password) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err = config.DB.QueryRow(query, customer.Name, customer.Email, customer.Phone, customer.Password).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		log.Println("❌ Error inserting customer:", err)
		// Check for unique constraint error just in case of race condition?
		return models.Customer{}, fmt.Errorf("failed to register customer: %w", err)
	}

	// Important: Clear password hash before returning the struct
	customer.Password = ""
	log.Printf("✅ Customer registered successfully with ID: %d", customer.ID)
	return customer, nil
}

func AuthenticateCustomer(email, password string) (string, error) {
	var customer models.Customer
	// Select required fields including password hash for checking
	query := "SELECT id, name, email, password, phone, created_at, updated_at FROM customers WHERE email=$1"
	err := config.DB.Get(&customer, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ Customer email not found: %s", email)
			return "", errors.New("invalid email or password") // Generic error
		}
		log.Printf("❌ Error fetching customer %s: %v", email, err)
		return "", fmt.Errorf("error fetching customer data: %w", err)
	}

	// Check password
	if !utils.CheckPasswordHash(password, customer.Password) {
		log.Printf("❌ Customer password mismatch for: %s", email)
		return "", errors.New("invalid email or password") // Generic error
	}

	// Generate token
	token, err := generateCustomerToken(customer.ID, customer.Email)
	if err != nil {
		log.Println("❌ Error generating customer token:", err)
		return "", errors.New("authentication failed")
	}

	log.Printf("✅ Authentication successful for customer: %s", email)
	return token, nil
}

// --- Token Generation (Keep as is, logic is specific) ---

func generateEmployeeToken(employeeID int, email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Consider making expiration configurable
	claims := jwt.MapClaims{
		"user_type":   "employee", // Add user type claim
		"employee_id": employeeID,
		"email":       email,
		"role":        role,
		"exp":         expirationTime.Unix(),
		"iss":         "car-rental-api", // Example issuer
		"iat":         time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JwtSecret))
	if err != nil {
		log.Println("❌ Error signing employee token:", err)
		return "", fmt.Errorf("failed to sign employee token: %w", err) // Wrap internal error
	}
	return tokenString, nil
}

func generateCustomerToken(customerID int, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Consider making expiration configurable
	claims := jwt.MapClaims{
		"user_type":   "customer", // Add user type claim
		"customer_id": customerID,
		"email":       email,
		"role":        "customer", // Explicit role for consistency
		"exp":         expirationTime.Unix(),
		"iss":         "car-rental-api", // Example issuer
		"iat":         time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JwtSecret))
	if err != nil {
		log.Println("❌ Error signing customer token:", err)
		return "", fmt.Errorf("failed to sign customer token: %w", err) // Wrap internal error
	}
	return tokenString, nil
}
