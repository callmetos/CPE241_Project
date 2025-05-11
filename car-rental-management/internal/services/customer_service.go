package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"car-rental-management/internal/utils"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
)

type CustomerFiltersWithPagination struct {
	Name          *string
	Email         *string
	Phone         *string
	Page          int
	Limit         int
	SortBy        string
	SortDirection string
}

type PaginatedCustomersResponse struct {
	Customers  []models.Customer `json:"customers"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

func GetCustomersPaginated(filters CustomerFiltersWithPagination) (PaginatedCustomersResponse, error) {
	var response PaginatedCustomersResponse
	response.Customers = []models.Customer{}

	queryBuilder := strings.Builder{}
	countQueryBuilder := strings.Builder{}
	args := []interface{}{}
	paramCount := 1

	queryBuilder.WriteString("SELECT id, name, email, phone, created_at, updated_at FROM customers")
	countQueryBuilder.WriteString("SELECT COUNT(*) FROM customers")

	var conditions []string
	if filters.Name != nil && *filters.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", paramCount))
		args = append(args, "%"+*filters.Name+"%")
		paramCount++
	}
	if filters.Email != nil && *filters.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", paramCount))
		args = append(args, "%"+*filters.Email+"%")
		paramCount++
	}
	if filters.Phone != nil && *filters.Phone != "" {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", paramCount))
		args = append(args, "%"+*filters.Phone+"%")
		paramCount++
	}

	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		queryBuilder.WriteString(whereClause)
		countQueryBuilder.WriteString(whereClause)
	}

	err := config.DB.QueryRow(countQueryBuilder.String(), args...).Scan(&response.TotalCount)
	if err != nil {
		log.Printf("Error counting customers: %v", err)
		return response, fmt.Errorf("failed to count customers: %w", err)
	}

	orderByClause := " ORDER BY id ASC"
	if filters.SortBy != "" {
		validSortByFields := map[string]bool{"id": true, "name": true, "email": true, "created_at": true}
		if validSortByFields[filters.SortBy] {
			sortDir := "ASC"
			if strings.ToUpper(filters.SortDirection) == "DESC" {
				sortDir = "DESC"
			}
			orderByClause = fmt.Sprintf(" ORDER BY %s %s, id %s", filters.SortBy, sortDir, sortDir)
		}
	}
	queryBuilder.WriteString(orderByClause)

	if filters.Limit <= 0 {
		filters.Limit = 10
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	offset := (filters.Page - 1) * filters.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount, paramCount+1))
	args = append(args, filters.Limit, offset)

	rows, err := config.DB.Queryx(queryBuilder.String(), args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, nil
		}
		log.Printf("Error fetching paginated customers: %v", err)
		return response, fmt.Errorf("failed to fetch customers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var customer models.Customer
		if err := rows.StructScan(&customer); err != nil {
			log.Printf("Error scanning customer: %v", err)
			continue
		}
		response.Customers = append(response.Customers, customer)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating customer rows: %v", err)
		return response, fmt.Errorf("error processing customer rows: %w", err)
	}

	response.Page = filters.Page
	response.Limit = filters.Limit
	if response.TotalCount > 0 && response.Limit > 0 {
		response.TotalPages = int(math.Ceil(float64(response.TotalCount) / float64(response.Limit)))
	} else {
		response.TotalPages = 0
	}
	log.Printf("Service: Fetched %d customers. Page: %d, Limit: %d, TotalItems: %d, TotalPages: %d", len(response.Customers), response.Page, response.Limit, response.TotalCount, response.TotalPages)
	return response, nil
}

func GetCustomers() ([]models.Customer, error) {
	paginatedFilters := CustomerFiltersWithPagination{
		Page:          1,
		Limit:         10000,
		SortBy:        "id",
		SortDirection: "ASC",
	}
	result, err := GetCustomersPaginated(paginatedFilters)
	if err != nil {
		return nil, err
	}
	return result.Customers, nil
}

func GetCustomerByID(id int) (models.Customer, error) {
	var customer models.Customer
	if id <= 0 {
		return models.Customer{}, errors.New("invalid customer ID")
	}
	query := "SELECT id, name, email, phone, created_at, updated_at FROM customers WHERE id=$1"
	err := config.DB.Get(&customer, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Customer{}, errors.New("customer not found")
		}
		return models.Customer{}, fmt.Errorf("failed to fetch customer: %w", err)
	}
	return customer, nil
}

func UpdateCustomer(customerID int, input models.UpdateCustomerByStaffInput) (models.Customer, error) {
	if customerID <= 0 {
		return models.Customer{}, errors.New("invalid customer ID for update")
	}
	if strings.TrimSpace(input.Name) == "" {
		return models.Customer{}, errors.New("customer name cannot be empty")
	}
	if !utils.IsValidEmail(input.Email) {
		return models.Customer{}, errors.New("invalid email format")
	}

	query := `UPDATE customers SET name=$1, email=$2, phone=$3 WHERE id=$4`
	result, err := config.DB.Exec(query, input.Name, input.Email, input.Phone, customerID)
	if err != nil {
		if strings.Contains(err.Error(), "customers_email_key") {
			return models.Customer{}, errors.New("email already exists for another customer")
		}
		return models.Customer{}, fmt.Errorf("failed to update customer: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.Customer{}, errors.New("customer not found for update")
	}
	updatedCustomer, fetchErr := GetCustomerByID(customerID)
	if fetchErr != nil {
		return models.Customer{}, errors.New("update succeeded but failed to fetch updated data")
	}
	return updatedCustomer, nil
}

func UpdateCustomerProfile(customerID int, input models.UpdateCustomerProfileInput) (models.Customer, error) {
	if customerID <= 0 {
		return models.Customer{}, errors.New("invalid customer ID")
	}
	if strings.TrimSpace(input.Name) == "" {
		return models.Customer{}, errors.New("customer name cannot be empty")
	}

	query := `UPDATE customers SET name=$1, phone=$2 WHERE id=$3`
	result, err := config.DB.Exec(query, input.Name, input.Phone, customerID)
	if err != nil {
		return models.Customer{}, fmt.Errorf("failed to update profile: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.Customer{}, errors.New("customer not found for profile update (or possibly deleted)")
	}
	updatedCustomer, fetchErr := GetCustomerByID(customerID)
	if fetchErr != nil {
		return models.Customer{}, errors.New("profile update succeeded but failed to retrieve updated profile")
	}
	return updatedCustomer, nil
}

func DeleteCustomer(customerID int) error {
	if customerID <= 0 {
		return errors.New("invalid customer ID for deletion")
	}
	result, err := config.DB.Exec("DELETE FROM customers WHERE id=$1", customerID)
	if err != nil {
		if strings.Contains(err.Error(), "rentals_customer_id_fkey") {
			return errors.New("cannot delete customer: they have associated rentals")
		}
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("customer not found for deletion")
	}
	return nil
}
