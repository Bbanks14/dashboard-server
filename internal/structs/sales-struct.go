package structs

import (
	"time"

	"github.com/google/uuid"
)

type Sale struct {
	SaleID      uuid.UUID `db:"sale_id" json:"sale_id"`
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	ProductID   uuid.UUID `db:"product_id" json:"product_id"`
	CustomerID  uuid.UUID `db:"customer_id" json:"customer_id"`
	Quanttity   int64     `db:"quantity" json:"quality"`
	SaleDate    time.Time `db:"sale_date" json:"sale_date"`
	TotalAmount float64   `db:"total_amount" json:"total_amount"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
