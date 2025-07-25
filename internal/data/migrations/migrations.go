// Use structured logging instead of log.Printf
package migrations

import (
	"context"
	"database/sql"
)

// internal/data/database/migrations.go or seeder.go
func SeedInitialData(ctx context.Context, db *sql.DB) error {
	// Check if data exists, insert if not
	// Similar to your JS insertInitialData function
}
