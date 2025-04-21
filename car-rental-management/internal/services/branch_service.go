package services

import (
	"car-rental-management/internal/config"
	"car-rental-management/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// CreateBranch adds a new branch after validation
func CreateBranch(branch models.Branch) (models.Branch, error) {
	log.Println("Attempting to create branch:", branch.Name)

	if strings.TrimSpace(branch.Name) == "" {
		return models.Branch{}, errors.New("branch name cannot be empty")
	}

	query := `INSERT INTO branches (name, address, phone) VALUES ($1, $2, $3)
			  RETURNING id, created_at, updated_at`

	var createdAt, updatedAt time.Time
	err := config.DB.QueryRow(query, branch.Name, branch.Address, branch.Phone).Scan(&branch.ID, &createdAt, &updatedAt)

	if err != nil {
		log.Printf("❌ Error inserting branch '%s': %v", branch.Name, err)
		if strings.Contains(err.Error(), "branches_name_key") {
			return models.Branch{}, errors.New("branch name already exists")
		}
		// Wrap the error
		return models.Branch{}, fmt.Errorf("failed to create branch in database: %w", err)
	}
	branch.CreatedAt = createdAt
	branch.UpdatedAt = updatedAt
	log.Printf("✅ Branch created successfully with ID: %d", branch.ID)
	return branch, nil
}

// GetBranches retrieves all branches
func GetBranches() ([]models.Branch, error) {
	log.Println("Fetching all branches")
	var branches []models.Branch
	query := "SELECT id, name, address, phone, created_at, updated_at FROM branches ORDER BY name"
	err := config.DB.Select(&branches, query)
	if err != nil {
		log.Println("❌ Error fetching branches:", err)
		// Wrap the error
		return nil, fmt.Errorf("failed to fetch branches from database: %w", err)
	}
	log.Printf("✅ Fetched %d branches successfully", len(branches))
	return branches, nil
}

// GetBranchByID retrieves a single branch by ID
func GetBranchByID(id int) (models.Branch, error) {
	if id <= 0 {
		return models.Branch{}, errors.New("invalid branch ID")
	}
	log.Println("Fetching branch by ID:", id)
	var branch models.Branch
	query := "SELECT id, name, address, phone, created_at, updated_at FROM branches WHERE id=$1"
	err := config.DB.Get(&branch, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("❌ Branch with ID %d not found.", id)
			return models.Branch{}, errors.New("branch not found") // Keep specific error
		}
		log.Printf("❌ Error fetching branch %d: %v", id, err)
		// Wrap the error
		return models.Branch{}, fmt.Errorf("failed to fetch branch %d from database: %w", id, err)
	}
	log.Printf("✅ Branch %d fetched successfully", id)
	return branch, nil
}

// UpdateBranch updates an existing branch
func UpdateBranch(branch models.Branch) (models.Branch, error) {
	log.Println("Attempting to update branch:", branch.ID)
	if branch.ID <= 0 {
		return models.Branch{}, errors.New("invalid branch ID for update")
	}
	if strings.TrimSpace(branch.Name) == "" {
		return models.Branch{}, errors.New("branch name cannot be empty")
	}

	query := `UPDATE branches SET name=:name, address=:address, phone=:phone WHERE id=:id`
	result, err := config.DB.NamedExec(query, branch)
	if err != nil {
		log.Printf("❌ Error updating branch %d: %v", branch.ID, err)
		if strings.Contains(err.Error(), "branches_name_key") {
			return models.Branch{}, errors.New("branch name already exists")
		}
		// Wrap the error
		return models.Branch{}, fmt.Errorf("failed to update branch %d in database: %w", branch.ID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Could not get rows affected for branch update %d: %v", branch.ID, err)
		// Return wrapped error as we don't know if update happened
		return models.Branch{}, fmt.Errorf("update query executed but failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return models.Branch{}, errors.New("branch not found for update") // Specific error
	}

	log.Printf("✅ Branch %d updated successfully", branch.ID)
	updatedBranch, fetchErr := GetBranchByID(branch.ID)
	if fetchErr != nil {
		log.Printf("⚠️ Failed to fetch updated branch data after update for ID %d: %v", branch.ID, fetchErr)
		// Return error indicating ambiguity
		return models.Branch{}, fmt.Errorf("branch update successful but failed to fetch updated data: %w", fetchErr)
	}
	return updatedBranch, nil
}

// DeleteBranch removes a branch after checking dependencies
func DeleteBranch(id int) error {
	log.Println("Attempting to delete branch:", id)
	if id <= 0 {
		return errors.New("invalid branch ID")
	}

	var carCount int
	// Use QueryRow for single value count check
	err := config.DB.QueryRow("SELECT COUNT(*) FROM cars WHERE branch_id=$1", id).Scan(&carCount)
	if err != nil {
		log.Printf("❌ Error checking cars in branch %d: %v", id, err)
		// Wrap the error
		return fmt.Errorf("failed to check dependencies before deleting branch: %w", err)
	}
	if carCount > 0 {
		log.Printf("⚠️ Cannot delete branch %d: it contains %d car(s)", id, carCount)
		return fmt.Errorf("cannot delete branch: %d car(s) assigned to it", carCount)
	}

	result, err := config.DB.Exec("DELETE FROM branches WHERE id=$1", id)
	if err != nil {
		log.Printf("❌ Error deleting branch %d: %v", id, err)
		// Check for FK constraint just in case (though check above should prevent)
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return errors.New("cannot delete branch due to existing dependencies (e.g., cars)")
		}
		// Wrap the error
		return fmt.Errorf("failed to delete branch %d from database: %w", id, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("⚠️ Could not get rows affected for branch delete %d: %v", id, err)
		return fmt.Errorf("delete query executed but failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("branch not found for deletion") // Specific error
	}
	log.Printf("✅ Branch %d deleted successfully", id)
	return nil
}
