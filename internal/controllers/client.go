package controllers

import (
	"net/http"
	"sync"

	"github.com/Bbanks14/dashboard-server/internal/models"

	"github.com/gin-gonic/gin"
)

// GetProducts gets product list using go routines
func GetProducts(c *gin.Context) {
	var products []models.ProductRepository

	if err := db.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	var wg sync.WaitGroup
	resultChan := make(chan ProductWithStat, len(products))

	for _, product := range products {
		wg.Add(1)

		go func(p Product) {
			defer wg.Done()

			var stats []ProductStat
			db.where("product_id ?", p.ID).Find(&stats)
			resultChan <- ProductWithStat{
				Product: p,
				Stats:   stats,
			}
		}(products)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var result []ProductWithStat
	for pw := range resultChan {
		result = append(result, pw)
	}

	c.JSON(http.StatusOK, result)
}

// GetCustomers retrieves all users with the role "user"
func GetCustomers(c *gin.Context) {
	var customers []models.User

	result := db.where("role = ?", "user").Find(&customers)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, customers)
}

// GetTransactions retrieves transactions with pagination, sorting and search
func GetTransactions(c *gin.Context) {
	// Create channels for concurrent operations
	var wg sync.WaitGroup
	wg.Add(2)

	transactionsChan := make(chan []Transaction, 1)
	totalChan := make(chan int64, 1)
	errorChan := make(chan error, 2)

	// Go routine for fetching transactions
	go func() {
		defer wg.Done()

		var transactions []Transaction
result := query.Offset(page * pageSize).Limit(pageSize).Find(&transactions)
		if result.Error != nil {
			errorChan <- result.Error
			return
		}
		
		transactionsChan <- transactions
	}()
	
	// Goroutine for counting total
	go func() {
		defer wg.Done()
		
		var count int64
		countQuery := db.Model(&Transaction{})
		if search != "" {
			countQuery = countQuery.Where("cost LIKE ? OR user_id LIKE ?", "%"+search+"%", "%"+search+"%")
		}
		
		result := countQuery.Count(&count)
		if result.Error != nil {
			errorChan <- result.Error
			return
		}
		
		totalChan <- count
	}()
	
	// Wait for goroutines to complete
	go func() {
		wg.Wait()
		close(transactionsChan)
		close(totalChan)
		close(errorChan)
	}()
	
	// Check for errors
	select {
	case err := <-errorChan:
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	default:
		// Continue if no errors
	}
	
	// Get results from channels
	transactions := <-transactionsChan
	total := <-totalChan
	
	// Return response
	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total": total,
	})
}

// getCountryIso3 returns the ISO3 code for a country
func getCountryIso3(country string) string {
	// Implementation of getCountryIso3 function would go here
	// This is a placeholder for the actual implementation
	return country // Simplified for this example
}

// GetGeography returns the count of users by country
func GetGeography(c *gin.Context) {
	var users []User
	
	result := db.Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": result.Error.Error()})
		return
	}
	
	// Process locations in a goroutine
	locationsChan := make(chan []Location, 1)
	errorChan := make(chan error, 1)
	
	go func() {
		mappedLocations := make(map[string]int)
		
		for _, user := range users {
			countryISO3 := getCountryIso3(user.Country)
			if _, exists := mappedLocations[countryISO3]; !exists {
				mappedLocations[countryISO3] = 1
			}
			mappedLocations[countryISO3]++
		}
		
		formattedLocations := make([]Location, 0, len(mappedLocations))
		for country, count := range mappedLocations {
			formattedLocations = append(formattedLocations, Location{
				ID:    country,
				Value: count,
			})
		}
		
		locationsChan <- formattedLocations
	}()
	
	// Get results from channel
	select {
	case err := <-errorChan:
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case locations := <-locationsChan:
		c.JSON(http.StatusOK, locations)
	}
}
