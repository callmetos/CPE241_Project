package services

import (
	"car-rental-management/config"
	"car-rental-management/models"
	"log"
)

// GetCustomers retrieves all customers from the database
func GetCustomers() ([]models.Customer, error) {
	var customers []models.Customer

	log.Println("🔍 Attempting to fetch customers from the database...")

	// Execute the database query
	err := config.DB.Select(&customers, "SELECT * FROM customers")

	if err != nil {
		log.Println("❌ Database query error:", err)
		return nil, err
	}

	log.Println("✅ Customers fetched successfully:", customers)
	return customers, nil
}

// UpdateCustomer updates customer details
func UpdateCustomer(customer models.Customer) error {
	log.Println("🔄 Updating customer:", customer.ID)

	_, err := config.DB.NamedExec(`
		UPDATE customers SET name=:name, email=:email, phone=:phone
		WHERE id=:id`, customer)

	if err != nil {
		log.Println("❌ Error updating customer:", err)
		return err
	}

	log.Println("✅ Customer updated successfully!")
	return nil
}

// DeleteCustomer removes a customer from the database
func DeleteCustomer(customerID int) error {
	log.Println("🗑 Deleting customer with ID:", customerID)

	_, err := config.DB.Exec("DELETE FROM customers WHERE id=$1", customerID)
	if err != nil {
		log.Println("❌ Error deleting customer:", err)
		return err
	}

	log.Println("✅ Customer deleted successfully!")
	return nil
}
