package models

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductRepository defines the methods you'll use to interact with the products table
type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	GetAll(ctx context.Context) ([]*Product, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}
