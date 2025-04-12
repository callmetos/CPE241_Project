package handlers

import (
	"car-rental-management/internal/models"
	"car-rental-management/internal/services"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CreateBranch handles POST /branches (requires admin/manager role via routing)
func CreateBranch(c *gin.Context) {
	var branch models.Branch
	if err := c.ShouldBindJSON(&branch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	createdBranch, err := services.CreateBranch(branch)
	if err != nil {
		log.Printf("❌ Handler: Error creating branch: %v", err)
		if err.Error() == "branch name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "empty") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch"})
		}
		return
	}
	c.JSON(http.StatusCreated, createdBranch)
}

// GetBranches handles GET /branches (public route)
func GetBranches(c *gin.Context) {
	branches, err := services.GetBranches()
	if err != nil {
		log.Printf("❌ Handler: Error getting branches: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branches"})
		return
	}
	c.JSON(http.StatusOK, branches)
}

// GetBranchByID handles GET /branches/:id (public route)
func GetBranchByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID format"})
		return
	}
	if id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID value"})
		return
	}

	branch, err := services.GetBranchByID(id)
	if err != nil {
		if errors.Is(err, errors.New("branch not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("❌ Handler: Error getting branch %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch details"})
		}
		return
	}
	c.JSON(http.StatusOK, branch)
}

// UpdateBranch handles PUT /branches/:id (requires admin/manager role via routing)
func UpdateBranch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID format"})
		return
	}
	if id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID value"})
		return
	}

	var branch models.Branch
	if err := c.ShouldBindJSON(&branch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	branch.ID = id // Set ID from URL param

	updatedBranch, err := services.UpdateBranch(branch)
	if err != nil {
		log.Printf("❌ Handler: Error updating branch %d: %v", id, err)
		if errors.Is(err, errors.New("branch not found for update")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "branch name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "empty") || strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branch"})
		}
		return
	}
	c.JSON(http.StatusOK, updatedBranch)
}

// DeleteBranch handles DELETE /branches/:id (requires admin/manager role via routing)
func DeleteBranch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID format"})
		return
	}
	if id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID value"})
		return
	}

	err = services.DeleteBranch(id)
	if err != nil {
		log.Printf("❌ Handler: Error deleting branch %d: %v", id, err)
		if errors.Is(err, errors.New("branch not found for deletion")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "cannot delete branch") { // Error from service check
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete branch"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Branch deleted successfully"})
}
