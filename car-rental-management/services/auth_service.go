package services

import (
	"car-rental-management/config"
	"car-rental-management/models"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterEmployee registers a new employee in the system
func RegisterEmployee(employee models.Employee) error {
	// Check if employee with the given email already exists
	var count int
	err := config.DB.Get(&count, "SELECT COUNT(*) FROM employees WHERE email=$1", employee.Email)
	if err != nil {
		log.Println("❌ Error checking existing email:", err)
		return err
	}

	if count > 0 {
		log.Println("⚠️ Email already exists:", employee.Email)
		return fmt.Errorf("email already exists")
	}

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(employee.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("❌ Error hashing password:", err)
		return err
	}
	employee.Password = string(hashedPassword)

	// Insert the new employee into the database
	query := `INSERT INTO employees (name, email, password, role) VALUES ($1, $2, $3, $4)`
	_, err = config.DB.Exec(query, employee.Name, employee.Email, employee.Password, employee.Role)
	if err != nil {
		log.Println("❌ Error registering employee:", err)
		return err
	}

	log.Println("✅ Employee registered successfully!")
	return nil
}

// AuthenticateEmployee checks the provided credentials and returns a JWT token if valid
func AuthenticateEmployee(email, password string) (string, error) {
	var employee models.Employee
	// Log the incoming request to verify the email
	log.Println("Authenticating employee with email:", email)

	// Use simple query without struct binding first to debug
	query := "SELECT id, name, email, password, role FROM employees WHERE email=$1"
	err := config.DB.QueryRow(query, email).Scan(
		&employee.ID,
		&employee.Name,
		&employee.Email,
		&employee.Password,
		&employee.Role,
	)

	if err != nil {
		log.Println("❌ Employee not found:", err)
		return "", fmt.Errorf("invalid credentials")
	}

	// Debugging logs
	log.Println("Found employee:", employee.Email, "with role:", employee.Role)
	log.Println("Comparing passwords...")

	// Compare the provided password with the stored hashed password
	err = bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(password))
	if err != nil {
		log.Println("❌ Password mismatch error:", err)
		return "", fmt.Errorf("invalid credentials")
	}

	log.Println("✅ Password match successful!")

	// Generate and return the JWT token if authentication succeeds
	token, err := generateToken(employee.Email, employee.Role)
	if err != nil {
		log.Println("❌ Error generating token:", err)
		return "", err
	}

	log.Println("✅ Authentication successful for:", employee.Email)
	return token, nil
}

// generateToken generates a JWT token for the authenticated employee
func generateToken(email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"email": email,
		"role":  role,
		"exp":   expirationTime.Unix(),
	}

	// Create a new JWT token with the claims and sign it using the secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JwtSecret))
	if err != nil {
		log.Println("❌ Error generating token:", err)
		return "", err
	}

	return tokenString, nil
}
