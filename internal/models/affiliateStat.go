package models

import (
	"github.com//google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AffiliateModel handles the database operations for the AffiliateStatSchema
type AffiliateModel struct {
	pool *pgxpool.Pool
}

// NewAffiliateModel creates a new instance of AffiliateModel
func NewAffiliateModel(dbPool *pgxpool.Pool) *AffiliateModel {
	return &AffiliateModel{
		pool: &dbPool,
	}
}
