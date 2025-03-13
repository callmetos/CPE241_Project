package models

type Customer struct {
	ID    int    `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
	Phone string `db:"phone" json:"phone"`
}
