package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"car-rental-management/internal/utils"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
)

func GetUsers() ([]models.Employee, error) {
	var users []models.Employee
	log.Println("🔍 Fetching users (employees)...")

	query := "SELECT id, name, email, role, created_at, updated_at FROM employees ORDER BY name"
	err := config.DB.Select(&users, query)
	if err != nil {
		log.Println("❌ Error fetching users (employees):", err)
		return nil, err
	}

	log.Printf("✅ Users (employees) fetched successfully! Count: %d", len(users))
	return users, nil
}

func CreateEmployeeByAdmin(input models.CreateEmployeeInput) (models.Employee, error) {
	log.Println("⚙️ Service: Admin creating employee:", input.Email)

	if strings.TrimSpace(input.Name) == "" {
		return models.Employee{}, errors.New("employee name cannot be empty")
	}
	if !utils.IsValidEmail(input.Email) {
		return models.Employee{}, errors.New("invalid email format")
	}
	if len(input.Password) < 6 {
		return models.Employee{}, errors.New("password must be at least 6 characters")
	}

	var count int
	countQuery := "SELECT COUNT(*) FROM employees WHERE email=$1"
	err := config.DB.Get(&count, countQuery, input.Email)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("❌ Service: DB error checking employee email existence for %s: %v", input.Email, err)
		return models.Employee{}, fmt.Errorf("database error checking email: %w", err)
	}
	if count > 0 {
		return models.Employee{}, errors.New("employee email already exists")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		log.Println("❌ Service: Error hashing employee password:", err)
		return models.Employee{}, errors.New("failed to secure password")
	}

	var createdEmployee models.Employee
	insertQuery := `
		INSERT INTO employees (name, email, password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, role, created_at, updated_at
	`
	err = config.DB.QueryRowx(insertQuery, input.Name, input.Email, hashedPassword, input.Role).StructScan(&createdEmployee)
	if err != nil {
		log.Println("❌ Service: Error inserting employee:", err)

		return models.Employee{}, fmt.Errorf("database error creating employee: %w", err)
	}

	log.Printf("✅ Service: Employee created successfully by admin with ID: %d", createdEmployee.ID)
	return createdEmployee, nil
}

func UpdateEmployeeByAdmin(id int, input models.UpdateEmployeeInput) (models.Employee, error) {
	log.Printf("⚙️ Service: Admin updating employee ID: %d", id)

	if id <= 0 {
		return models.Employee{}, errors.New("invalid employee ID")
	}

	if strings.TrimSpace(input.Name) == "" {
		return models.Employee{}, errors.New("employee name cannot be empty")
	}
	if !utils.IsValidEmail(input.Email) {
		return models.Employee{}, errors.New("invalid email format")
	}

	query := `
		UPDATE employees SET name=$1, email=$2, role=$3
		WHERE id=$4
		RETURNING id, name, email, role, created_at, updated_at
	`
	var updatedEmployee models.Employee
	err := config.DB.QueryRowx(query, input.Name, input.Email, input.Role, id).StructScan(&updatedEmployee)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Employee{}, errors.New("employee not found for update")
		}

		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {

			if pgErr.Constraint == "employees_email_key" {
				return models.Employee{}, errors.New("email already exists for another employee")
			}
		}
		log.Printf("❌ Service: Error updating employee %d: %v", id, err)
		return models.Employee{}, fmt.Errorf("database error updating employee: %w", err)
	}

	log.Printf("✅ Service: Employee %d updated successfully by admin.", id)
	return updatedEmployee, nil
}

func DeleteEmployeeByAdmin(id int) error {
	log.Printf("⚙️ Service: Admin deleting employee ID: %d", id)

	if id <= 0 {
		return errors.New("invalid employee ID")
	}

	result, err := config.DB.Exec("DELETE FROM employees WHERE id=$1", id)
	if err != nil {
		log.Printf("❌ Service: Error deleting employee %d: %v", id, err)

		return fmt.Errorf("database error deleting employee: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {

		log.Printf("⚠️ Service: Could not get rows affected for employee delete %d: %v", id, err)
	}
	if rowsAffected == 0 {
		return errors.New("employee not found for deletion")
	}

	log.Printf("✅ Service: Employee %d deleted successfully by admin.", id)
	return nil
}
