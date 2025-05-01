package models

import "time"

type Payment struct {
	ID          int       `db:"id" json:"id"`
	RentalID    int       `db:"rental_id" json:"rental_id"` // Removed binding:"required" as it might be created later
	Amount      float64   `db:"amount" json:"amount"`       // Amount might be calculated/set later
	PaymentDate time.Time `db:"payment_date" json:"payment_date"`
	// *** เพิ่ม 'Pending Verification' ใน CHECK constraint (ต้องแก้ไขใน SQL ด้วย) ***
	PaymentStatus        string  `db:"payment_status" json:"payment_status"`                   // Possible: Pending, Paid, Failed, Refunded, Pending Verification
	PaymentMethod        *string `db:"payment_method" json:"payment_method"`                   // e.g., "Bank Transfer", "QR Code"
	RecordedByEmployeeID *int    `db:"recorded_by_employee_id" json:"recorded_by_employee_id"` // Null if paid by customer online
	TransactionID        *string `db:"transaction_id" json:"transaction_id"`                   // Optional gateway transaction ID
	// *** เพิ่ม Field เก็บ URL สลิป ***
	SlipURL   *string   `db:"slip_url" json:"slip_url"` // Optional: Store slip file path/URL
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Input struct สำหรับ Admin/Staff บันทึก Payment (เหมือนเดิม)
type RecordPaymentInput struct {
	Amount        float64 `json:"amount" binding:"required,gte=0"`
	PaymentStatus string  `json:"payment_status" binding:"required,oneof=Paid Failed Refunded"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	TransactionID *string `json:"transaction_id"`
}

// ไม่จำเป็นต้องมี Input struct สำหรับ Upload Slip ใน Model โดยตรง
// เพราะข้อมูลหลักคือไฟล์ และ rentalId มาจาก path parameter
