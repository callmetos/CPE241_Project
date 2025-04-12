package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models" // Use utils if needed (e.g., email validation)
	"database/sql"
	"errors"
	"log"
	"strings"
	// Added for model fields
)

// GetCustomers retrieves all customers (excluding passwords)
func GetCustomers() ([]models.Customer, error) {
	var customers []models.Customer
	log.Println("🔍 Fetching customers...")
	query := "SELECT id, name, email, phone, created_at, updated_at FROM customers ORDER BY name" // Omit password
	err := config.DB.Select(&customers, query)
	if err != nil {
		log.Println("❌ Error fetching customers:", err)
		return nil, errors.New("failed to fetch customers")
	}
	log.Printf("✅ Customers fetched successfully! Count: %d", len(customers))
	return customers, nil
}

// GetCustomerByID retrieves a single customer by ID (excluding password)
func GetCustomerByID(id int) (models.Customer, error) {
	var customer models.Customer
	log.Println("🔍 Fetching customer by ID:", id)
	query := "SELECT id, name, email, phone, created_at, updated_at FROM customers WHERE id=$1" // Omit password
	err := config.DB.Get(&customer, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Customer{}, errors.New("customer not found")
		}
		log.Printf("❌ Error fetching customer %d: %v", id, err)
		return models.Customer{}, errors.New("failed to fetch customer")
	}
	log.Printf("✅ Customer %d fetched successfully!", id)
	return customer, nil
}

// UpdateCustomer updates customer details (by staff, cannot update password)
func UpdateCustomer(customer models.Customer) (models.Customer, error) {
	log.Println("🔄 Updating customer by staff:", customer.ID)
	// Validation
	if customer.ID <= 0 {
		return models.Customer{}, errors.New("invalid customer ID for update")
	}
	if strings.TrimSpace(customer.Name) == "" {
		return models.Customer{}, errors.New("customer name cannot be empty")
	}
	if !isValidEmail(customer.Email) {
		return models.Customer{}, errors.New("invalid email format")
	} // Use util if defined there

	query := `UPDATE customers SET name=:name, email=:email, phone=:phone WHERE id=:id` // updated_at handled by trigger
	result, err := config.DB.NamedExec(query, customer)
	if err != nil {
		log.Println("❌ Error updating customer by staff:", err)
		// Check for unique constraint violation on email
		if strings.Contains(err.Error(), "customers_email_key") {
			return models.Customer{}, errors.New("email already exists for another customer")
		}
		return models.Customer{}, errors.New("failed to update customer")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.Customer{}, errors.New("customer not found for update")
	}
	log.Println("✅ Customer updated successfully by staff!")
	// Fetch updated data
	updatedCustomer, fetchErr := GetCustomerByID(customer.ID)
	if fetchErr != nil {
		log.Printf("⚠️ Failed to fetch updated customer data after update for ID %d: %v", customer.ID, fetchErr)
		customer.Password = "" // Ensure password not returned
		return customer, nil   // Return input data as fallback
	}
	return updatedCustomer, nil
}

// UpdateCustomerProfile updates limited customer details (by customer themselves)
func UpdateCustomerProfile(customerID int, input models.UpdateCustomerProfileInput) (models.Customer, error) {
	log.Println("🔄 Updating customer profile by self:", customerID)
	// Validation
	if strings.TrimSpace(input.Name) == "" {
		return models.Customer{}, errors.New("customer name cannot be empty")
	}

	query := `UPDATE customers SET name=$1, phone=$2 WHERE id=$3` // Only allow updating name/phone
	result, err := config.DB.Exec(query, input.Name, input.Phone, customerID)
	if err != nil {
		log.Println("❌ Error updating customer profile:", err)
		return models.Customer{}, errors.New("failed to update profile")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.Customer{}, errors.New("customer not found for profile update")
	}
	log.Println("✅ Customer profile updated successfully!")
	updatedCustomer, fetchErr := GetCustomerByID(customerID) // Fetch again to get latest data
	if fetchErr != nil {
		log.Printf("⚠️ Failed to fetch updated customer data after profile update for ID %d: %v", customerID, fetchErr)
		// Cannot return input as it's different struct, return error or empty
		return models.Customer{}, errors.New("failed to retrieve updated profile")
	}
	return updatedCustomer, nil
}

// DeleteCustomer removes a customer (by staff)
func DeleteCustomer(customerID int) error {
	log.Println("🗑 Deleting customer by staff:", customerID)
	// Check dependencies first (active rentals? DB constraint ON DELETE RESTRICT handles this)
	result, err := config.DB.Exec("DELETE FROM customers WHERE id=$1", customerID)
	if err != nil {
		log.Println("❌ Error deleting customer:", err)
		// Check for FK violation
		if strings.Contains(err.Error(), "violates foreign key constraint") && strings.Contains(err.Error(), "rentals_customer_id_fkey") {
			return errors.New("cannot delete customer: they have associated rentals")
		}
		return errors.New("failed to delete customer")
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("customer not found for deletion")
	}
	log.Println("✅ Customer deleted successfully by staff!")
	return nil
}
