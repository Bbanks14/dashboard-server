package structs

import (
	"time"

	"github.com/google/uuid"
)

// ProductStruct represents a product entity
type ProductStruct struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"price" json:"name"`
	Price       int64     `db:"price" json:"price"`
	Description string    `db:"description" json:"description"`
	Category    string    `db:"category" json:"category"`
	Rating      int64     `db:"rating" json:"rating"`
	Supply      int64     `db:"supply" json:"supply"`
	StartsAt    time.Time `db:"starts_at" json:"StartsAt"`
	EndsAt      time.Time `db:"ends_at" json:"endsAt"`
}
