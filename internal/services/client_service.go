package services

import (
	"errors"

	"github.com/Bbanks14/dashboard-server/internal/repositories"
)

// ClientServiceInterface defines the contract for client business operations
type ClientServiceInterface interface {
	GetClients(page, pageSize int, sort, search string) (interface{}, int, error)
	GetProducts(page, pageSize int, sort, search string) (interface{}, int, error)
	GetCustomers(page, pageSize int, sort, search string) (interface{}, int, error)
	GetTransactions(page, pageSize int, sort, search string) (interface{}, int, error)
	GetGeography() (interface{}, error)
}

// ClientService handles business logic for client operations
type ClientService struct {
	repo repositories.ClientRepositoryInterface
}

// NewClientService creates a new client service instance
func NewClientService(repo repositories.ClientRepositoryInterface) ClientServiceInterface {
	return &ClientService{
		repo: repo,
	}
}

// GetClients retrieves clients with pagination and filtering
func (s *ClientService) GetClients(page, pageSize int, sort, search string) (interface{}, int, error) {
	// Business logic validation
	if err := s.validatePaginationParams(&page, &pageSize); err != nil {
		return nil, 0, err
	}

	if err := s.validateSortParam(sort); err != nil {
		return nil, 0, err
	}

	clients, totalCount, err := s.repo.GetClients(page, pageSize, sort, search)
	if err != nil {
		return nil, 0, err
	}

	// Add any post-processing business logic here if needed
	// For example: filtering sensitive data, calculating derived fields, etc.

	return clients, totalCount, nil
}

// GetProducts retrieves products with pagination and filtering
func (s *ClientService) GetProducts(page, pageSize int, sort, search string) (interface{}, int, error) {
	// Business logic validation
	if err := s.validatePaginationParams(&page, &pageSize); err != nil {
		return nil, 0, err
	}

	if err := s.validateSortParam(sort); err != nil {
		return nil, 0, err
	}

	products, totalCount, err := s.repo.GetProducts(page, pageSize, sort, search)
	if err != nil {
		return nil, 0, err
	}

	// Add any business logic for product processing here
	// For example: calculate discounts, apply business rules, etc.

	return products, totalCount, nil
}

// GetCustomers retrieves customers/users with pagination and filtering
func (s *ClientService) GetCustomers(page, pageSize int, sort, search string) (interface{}, int, error) {
	// Business logic validation
	if err := s.validatePaginationParams(&page, &pageSize); err != nil {
		return nil, 0, err
	}

	if err := s.validateSortParam(sort); err != nil {
		return nil, 0, err
	}

	users, totalCount, err := s.repo.GetUsers(page, pageSize, sort, search)
	if err != nil {
		return nil, 0, err
	}

	// Add any business logic for customer processing here
	// For example: mask sensitive information, calculate customer metrics, etc.

	return users, totalCount, nil
}

// GetTransactions retrieves transactions with pagination and filtering
func (s *ClientService) GetTransactions(page, pageSize int, sort, search string) (interface{}, int, error) {
	// Business logic validation
	if err := s.validatePaginationParams(&page, &pageSize); err != nil {
		return nil, 0, err
	}

	if err := s.validateSortParam(sort); err != nil {
		return nil, 0, err
	}

	transactions, totalCount, err := s.repo.GetTransactions(page, pageSize, sort, search)
	if err != nil {
		return nil, 0, err
	}

	// Add any business logic for transaction processing here
	// For example: calculate totals, apply business rules, audit logging, etc.

	return transactions, totalCount, nil
}

// GetGeography retrieves user count by location for geography charts
func (s *ClientService) GetGeography() (interface{}, error) {
	usersByCountry, err := s.repo.GetUsersByLocation()
	if err != nil {
		return nil, err
	}

	// Add any business logic for geography data processing here
	// For example: data aggregation, filtering, formatting for charts, etc.

	return usersByCountry, nil
}

// validatePaginationParams validates and normalizes pagination parameters
func (s *ClientService) validatePaginationParams(page, pageSize *int) error {
	if *page < 1 {
		*page = 1
	}

	if *pageSize < 1 {
		*pageSize = 10
	}

	if *pageSize > 100 {
		return errors.New("page size cannot exceed 100")
	}

	return nil
}

// validateSortParam validates sort parameter
func (s *ClientService) validateSortParam(sort string) error {
	// Define allowed sort fields
	allowedSortFields := map[string]bool{
		"id":         true,
		"name":       true,
		"email":      true,
		"created_at": true,
		"updated_at": true,
		// Add more allowed fields as needed
	}

	if sort != "" && !allowedSortFields[sort] {
		return errors.New("invalid sort field")
	}

	return nil
}
