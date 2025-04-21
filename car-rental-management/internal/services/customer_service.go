package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"car-rental-management/internal/utils" // Use utils for validation
	"database/sql"
	"errors"
	"fmt" // For error wrapping
	"log"
	"strings"
)

// GetCustomers retrieves all customers (excluding passwords)
func GetCustomers() ([]models.Customer, error) {
	var customers []models.Customer
	log.Println("🔍 Fetching customers...")
	query := "SELECT id, name, email, phone, created_at, updated_at FROM customers ORDER BY name" // Omit password
	err := config.DB.Select(&customers, query)
	if err != nil {
		log.Println("❌ Error fetching customers:", err)
		return nil, fmt.Errorf("failed to fetch customers: %w", err) // Wrap error
	}
	log.Printf("✅ Customers fetched successfully! Count: %d", len(customers))
	return customers, nil
}

// GetCustomerByID retrieves a single customer by ID (excluding password)
func GetCustomerByID(id int) (models.Customer, error) {
	var customer models.Customer
	log.Println("🔍 Fetching customer by ID:", id)
	if id <= 0 {
		return models.Customer{}, errors.New("invalid customer ID")
	}
	query := "SELECT id, name, email, phone, created_at, updated_at FROM customers WHERE id=$1" // Omit password
	err := config.DB.Get(&customer, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Customer{}, errors.New("customer not found") // Specific error for not found
		}
		log.Printf("❌ Error fetching customer %d: %v", id, err)
		return models.Customer{}, fmt.Errorf("failed to fetch customer: %w", err) // Wrap error
	}
	log.Printf("✅ Customer %d fetched successfully!", id)
	return customer, nil
}

// UpdateCustomer updates customer details (by staff, cannot update password)
// Now uses UpdateCustomerByStaffInput
func UpdateCustomer(customerID int, input models.UpdateCustomerByStaffInput) (models.Customer, error) {
	log.Println("🔄 Updating customer by staff:", customerID)
	// Validation
	if customerID <= 0 {
		return models.Customer{}, errors.New("invalid customer ID for update")
	}
	if strings.TrimSpace(input.Name) == "" {
		return models.Customer{}, errors.New("customer name cannot be empty")
	}
	if !utils.IsValidEmail(input.Email) { // Use util
		return models.Customer{}, errors.New("invalid email format")
	}

	query := `UPDATE customers SET name=$1, email=$2, phone=$3 WHERE id=$4` // updated_at handled by trigger
	result, err := config.DB.Exec(query, input.Name, input.Email, input.Phone, customerID)
	if err != nil {
		log.Println("❌ Error updating customer by staff:", err)
		// Check for unique constraint violation on email
		if strings.Contains(err.Error(), "customers_email_key") { // Simple check
			return models.Customer{}, errors.New("email already exists for another customer")
		}
		// Consider using pq error code check for more robustness:
		// if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { ... }
		return models.Customer{}, fmt.Errorf("failed to update customer: %w", err) // Wrap error
	}
	rowsAffected, _ := result.RowsAffected() // Ignore error on RowsAffected if Exec was ok
	if rowsAffected == 0 {
		return models.Customer{}, errors.New("customer not found for update") // Specific error
	}
	log.Println("✅ Customer updated successfully by staff!")
	// Fetch updated data
	updatedCustomer, fetchErr := GetCustomerByID(customerID)
	if fetchErr != nil {
		// Don't wrap fetchErr here, as the update itself succeeded.
		log.Printf("⚠️ Failed to fetch updated customer data after update for ID %d: %v", customerID, fetchErr)
		// Return a manually constructed struct or error
		return models.Customer{}, errors.New("update succeeded but failed to fetch updated data")
	}
	return updatedCustomer, nil
}

// UpdateCustomerProfile updates limited customer details (by customer themselves)
func UpdateCustomerProfile(customerID int, input models.UpdateCustomerProfileInput) (models.Customer, error) {
	log.Println("🔄 Updating customer profile by self:", customerID)
	if customerID <= 0 {
		return models.Customer{}, errors.New("invalid customer ID") // Should be caught by middleware/handler
	}
	// Validation
	if strings.TrimSpace(input.Name) == "" {
		return models.Customer{}, errors.New("customer name cannot be empty")
	}

	query := `UPDATE customers SET name=$1, phone=$2 WHERE id=$3` // Only allow updating name/phone
	result, err := config.DB.Exec(query, input.Name, input.Phone, customerID)
	if err != nil {
		log.Println("❌ Error updating customer profile:", err)
		return models.Customer{}, fmt.Errorf("failed to update profile: %w", err) // Wrap error
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// If token is valid, this should indicate the user was deleted between login and update.
		return models.Customer{}, errors.New("customer not found for profile update (or possibly deleted)")
	}
	log.Println("✅ Customer profile updated successfully!")
	updatedCustomer, fetchErr := GetCustomerByID(customerID) // Fetch again to get latest data
	if fetchErr != nil {
		log.Printf("⚠️ Failed to fetch updated customer data after profile update for ID %d: %v", customerID, fetchErr)
		return models.Customer{}, errors.New("profile update succeeded but failed to retrieve updated profile")
	}
	return updatedCustomer, nil
}

// DeleteCustomer removes a customer (by staff)
func DeleteCustomer(customerID int) error {
	log.Println("🗑 Deleting customer by staff:", customerID)
	if customerID <= 0 {
		return errors.New("invalid customer ID for deletion")
	}
	// Check dependencies first (active rentals? DB constraint ON DELETE RESTRICT handles this)
	result, err := config.DB.Exec("DELETE FROM customers WHERE id=$1", customerID)
	if err != nil {
		log.Println("❌ Error deleting customer:", err)
		// Check for FK violation
		if strings.Contains(err.Error(), "violates foreign key constraint") && strings.Contains(err.Error(), "rentals_customer_id_fkey") {
			// Consider checking pq error code "23503" for FK violation
			return errors.New("cannot delete customer: they have associated rentals")
		}
		// Wrap other errors
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("customer not found for deletion") // Specific error
	}
	log.Println("✅ Customer deleted successfully by staff!")
	return nil
}

// Removed local isValidEmail function
