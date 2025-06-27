package repositories

import (
	"github.com/Bbanks14/dashboard-server/internal/data/database"
)

// ClientRepositoryInterface defines the contract for client data operations
type ClientRepositoryInterface interface {
	GetClients(page, pageSize int, sort, search string) (interface{}, int, error)
	GetProducts(page, pageSize int, sort, search string) (interface{}, int, error)
	GetUsers(page, pageSize int, sort, search string) (interface{}, int, error)
	GetTransactions(page, pageSize int, sort, search string) (interface{}, int, error)
	GetUsersByLocation() (interface{}, error)
}

// ClientRepository handles data access for client operations
type ClientRepository struct {
	db *database.Database
}

// NewClientRepository creates a new client repository instance
func NewClientRepository(db *database.Database) ClientRepositoryInterface {
	return &ClientRepository{
		db: db,
	}
}

// GetClients retrieves clients from the database
func (r *ClientRepository) GetClients(page, pageSize int, sort, search string) (interface{}, int, error) {
	return r.db.GetClients(page, pageSize, sort, search)
}

// GetProducts retrieves products from the database
func (r *ClientRepository) GetProducts(page, pageSize int, sort, search string) (interface{}, int, error) {
	return r.db.GetProducts(page, pageSize, sort, search)
}

// GetUsers retrieves users from the database
func (r *ClientRepository) GetUsers(page, pageSize int, sort, search string) (interface{}, int, error) {
	return r.db.GetUsers(page, pageSize, sort, search)
}

// GetTransactions retrieves transactions from the database
func (r *ClientRepository) GetTransactions(page, pageSize int, sort, search string) (interface{}, int, error) {
	return r.db.GetTransactions(page, pageSize, sort, search)
}

// GetUsersByLocation retrieves user count by location from the database
func (r *ClientRepository) GetUsersByLocation() (interface{}, error) {
	return r.db.GetUsersByLocation()
}
