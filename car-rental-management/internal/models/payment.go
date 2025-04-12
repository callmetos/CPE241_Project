package models

import "time"

type Payment struct {
	ID                   int       `db:"id" json:"id"`
	RentalID             int       `db:"rental_id" json:"rental_id" binding:"required"`
	Amount               float64   `db:"amount" json:"amount" binding:"required,gte=0"`
	PaymentDate          time.Time `db:"payment_date" json:"payment_date"`
	PaymentStatus        string    `db:"payment_status" json:"payment_status" binding:"required,oneof=Pending Paid Failed Refunded"`
	PaymentMethod        *string   `db:"payment_method" json:"payment_method"`                   // Pointer for nullable
	RecordedByEmployeeID *int      `db:"recorded_by_employee_id" json:"recorded_by_employee_id"` // Pointer for nullable
	TransactionID        *string   `db:"transaction_id" json:"transaction_id"`                   // Pointer for nullable
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}

// Input struct specifically for recording payment by staff
type RecordPaymentInput struct {
	Amount        float64 `json:"amount" binding:"required,gte=0"`
	PaymentStatus string  `json:"payment_status" binding:"required,oneof=Paid Failed Refunded"` // Staff likely records final states
	PaymentMethod string  `json:"payment_method" binding:"required"`
	TransactionID *string `json:"transaction_id"` // Optional gateway transaction ID
}
