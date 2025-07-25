package services

import (
	"errors"
	"fmt"

	"github.com/Bbanks14/dashboard-server/internal/models"
	"github.com/Bbanks14/dashboard-server/internal/repository"
)

// ClientServiceInterface defines the contract for client business operations.
type ClientServiceInterface interface {
	GetClients(page, pageSize int, sort, search string) ([]models.Client, int, error)
	GetProducts(page, pageSize int, sort, search string) ([]models.Product, int, error)
	GetUsers(page, pageSize int, sort, search string) ([]models.User, int, error)
	GetTransactions(page, pageSize (map[string]int, error))
}

// ClientService handles business logic for client operations.
type ClientService struct {
	repo repository.ClientRepositoryInterface
}

// NewClientService creates a new client service instance.
func NewClientService(repo repository.ClientRepositoryInterface) ClientServiceInterface {
	return &ClientService{
		repo: repo,
	}
}

// GetClients retrieves clients with pagination and filtering.
func (s *ClientService) GetClients(page, pageSize int, sort, search string) ([]models.Client, int, error) {
	if err := s.validatePaginationParams(page, pageSize); err != nil {
		return nil, 0, err
	}
	if err := s.validateSortParam(sort); err != nil {
		return nil, 0, err
	}
	clients, totalCount, err := s.repo.GetClients(page, pageSize, sort, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get clients: %w", err)
	}
	// Add any post-processing business logic here if needed
	return clients, totalCount, nil
}

// GetProducts retrieves products with pagination and filtering.
func (s *ClientService) GetProducts(page, pageSize int, sort, search string) ([]models.Product, int, error) {
	if err := s.validatePaginationParams(page, pageSize); err != nil {
		return nil, 0, err
	}
	if err := s.validateSortParam(sort); err != nil {
		return nil, 0, err
	}
	products, totalCount, err := s.repo.GetProducts(page, pageSize, sort, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}
	// Add any business logic for product processing here
	return products, totalCount, nil
}

// GetUsers retrieves users with pagination and filtering.
func (s *ClientService) GetUsers(page, pageSize int, sort, search string) ([]models.User, int, error) {
	if err := s.validatePaginationParams(page, pageSize); err != nil {
		return nil, 0, err
	}
	if err := s.validateSortParam(sort); err != nil {
		return nil, 0, err
	}
	users, totalCount, err := s.repo.GetUsers(page, pageSize, sort, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}
	// Add any business logic for user processing here
	return users, totalCount, nil
}

// GetTransactions retrieves transactions with pagination and filtering.
func (s *ClientService) GetTransactions(page, pageSize int, sort, search string) ([]models.Transaction, int, error) {
	if err := s.validatePaginationParams(page, pageSize); err != nil {
		return nil, 0, err
	}
	if err := s.validateSortParam(sort); err != nil {
		return nil, 0, err
	}
	transactions, totalCount, err := s.repo.GetTransactions(page, pageSize, sort, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}
	// Add any business logic for transaction processing here
	return transactions, totalCount, nil
}

// GetGeography retrieves user count by location for geography charts.
func (s *ClientService) GetGeography() (map[string]int, error) {
	usersByCountry, err := s.repo.GetUsersByLocation()
	if err != nil {
		return nil, fmt.Errorf("failed to get geography data: %w", err)
	}
	// Add any business logic for geography data processing here
	return usersByCountry, nil
}

// validatePaginationParams validates pagination parameters.
func (s *ClientService) validatePaginationParams(page, pageSize int) error {
	if page < 1 {
		return errors.New("page must be at least 1")
	}
	if pageSize < 1 {
		return errors.New("pageSize must be at least 1")
	}
	if pageSize > 100 {
		return errors.New("pageSize cannot exceed 100")
	}
	return nil
}

// validateSortParam validates the sort parameter.
func (s *ClientService) validateSortParam(sort string) error {
	// Define allowed sort fields. Update this list if the schema changes.
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
