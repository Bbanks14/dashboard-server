package structs

import (
	"github.com/Bbanks14/dashboard-server/pkg/helpers"
)

// ProductStruct represents a product entity
type ProductStruct struct {
	Name        string             `json:"name"`
	Price       int64              `json:"price"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	Rating      int64              `json:"rating"`
	Supply      int64              `json:"supply"`
	StartsAt    helpers.customTime `json:"starts_at"`
	EndsAt      helpers.customTime `json:"ends_at"`
}
