package models

import "time"

// Review struct (ยังคงเดิม)
type Review struct {
	ID         int       `db:"id" json:"id"`
	CustomerID int       `db:"customer_id" json:"customer_id"`
	RentalID   int       `db:"rental_id" json:"rental_id"`
	Rating     int       `db:"rating" json:"rating" binding:"required,min=1,max=5"`
	Comment    *string   `db:"comment" json:"comment"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// CreateReviewInput struct (ยังคงเดิม)
type CreateReviewInput struct {
	Rating  int     `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment"`
}

// AdminReviewView struct (ยังคงเดิม - สำหรับ Admin ดูรีวิว)
type AdminReviewView struct {
	ID              int       `db:"id" json:"id"`
	RentalID        int       `db:"rental_id" json:"rental_id"`
	Rating          int       `db:"rating" json:"rating"`
	Comment         *string   `db:"comment" json:"comment"`
	ReviewCreatedAt time.Time `db:"review_created_at" json:"review_created_at"`
	CustomerID      int       `db:"customer_id" json:"customer_id"`
	CustomerName    *string   `db:"customer_name" json:"customer_name"`
	CarID           int       `db:"car_id" json:"car_id"`
	CarBrand        *string   `db:"car_brand" json:"car_brand"`
	CarModel        *string   `db:"car_model" json:"car_model"`
}

// ReviewFiltersWithPagination struct (ยังคงเดิม - สำหรับ Admin กรองรีวิว)
type ReviewFiltersWithPagination struct {
	Rating        *int
	CustomerID    *int
	CarID         *int
	Keyword       *string
	Page          int
	Limit         int
	SortBy        string
	SortDirection string
}

// PaginatedAdminReviewsResponse struct (ยังคงเดิม - สำหรับ response รีวิวของ Admin)
type PaginatedAdminReviewsResponse struct {
	Reviews    []AdminReviewView `json:"reviews"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}
