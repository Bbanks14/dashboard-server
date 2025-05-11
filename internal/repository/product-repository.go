package repository

import (
	"context"
	"fmt"

	"github.com/Bbanks14/dashboard-server/internal/structs"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPGXProductRepo constructs a new repository over a pgxpool
func NewPGXProductRepo(structs.pgxProductRepoStruct.db *pgxpool.Pool) ProductRepository {
	return &structs.pgxProductRepoStruct{db: db}
}

func (r *pgxProductRepo) Create(ctx context.Context, p *Product) error {
    // Generate a new UUID for the product
    p.ID = uuid.New()
    query := `
        INSERT INTO products (id, name, price_cents, released_at, discontinued_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    _, err := r.db.Exec(ctx, query,
        p.ID, p.Name, p.PriceCents, p.ReleasedAt, p.DiscontinuedAt,
    )
    return err
}

func (r *pgxProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
    const query = `
        SELECT id, name, price_cents, released_at, discontinued_at
        FROM products
        WHERE id = $1
    `
    row := r.db.QueryRow(ctx, query, id)
    var p Product
    err := row.Scan(&p.ID, &p.Name, &p.PriceCents, &p.ReleasedAt, &p.DiscontinuedAt)
    if err != nil {
        return nil, fmt.Errorf("fetch product by id: %w", err)
    }
    return &p, nil
}

// GetAll returns all products in the table.
func (r *pgxProductRepo) GetAll(ctx context.Context) ([]*Product, error) {
    const query = `
        SELECT id, name, price_cents, released_at, discontinued_at
        FROM products
    `
    rows, err := r.db.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query all products: %w", err)
    }
    defer rows.Close()

    var products []*Product
    for rows.Next() {
        var p Product
        err := rows.Scan(
            &p.ID,
            &p.Name,
            &p.PriceCents,
            &p.ReleasedAt,
            &p.DiscontinuedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("scan product: %w", err)
        }
        products = append(products, &p)
    }
    if rows.Err() != nil {
        return nil, fmt.Errorf("iterate products: %w", rows.Err())
    }

    return products, nil
}
func (r *pgxProductRepo) Update(ctx context.Context, p *Product) error {
    const query = `
        UPDATE products
        SET name = $2, price_cents = $3, released_at = $4, discontinued_at = $5
        WHERE id = $1
    `
    _, err := r.db.Exec(ctx, query,
        p.ID, p.Name, p.PriceCents, p.ReleasedAt, p.DiscontinuedAt,
    )
    return err
}

func (r *pgxProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
    const query = `DELETE FROM products WHERE id = $1`
    _, err := r.db.Exec(ctx, query, id)
    return err
}
