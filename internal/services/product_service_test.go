package services_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ybds/internal/models/product"
	"github.com/ybds/internal/services"
)

// TestProductService tests the ProductService functionality
func TestProductService(t *testing.T) {
	// This is an integration test that would require a database
	// In a real-world scenario, you would use a test database or mock the database
	t.Skip("Skipping integration test")
}

// TestProductResult tests the ProductResult struct
func TestProductResult(t *testing.T) {
	// Create a ProductResult
	productID := uuid.New()
	result := services.ProductResult{
		Success:   true,
		Message:   "Product created successfully",
		ProductID: productID,
		Name:      "Test Product",
		SKU:       "TEST-123",
	}

	// Test the fields
	assert.True(t, result.Success)
	assert.Equal(t, "Product created successfully", result.Message)
	assert.Equal(t, productID, result.ProductID)
	assert.Equal(t, "Test Product", result.Name)
	assert.Equal(t, "TEST-123", result.SKU)
}

// TestInventoryResult tests the InventoryResult struct
func TestInventoryResult(t *testing.T) {
	// Create an InventoryResult
	inventoryID := uuid.New()
	productID := uuid.New()
	result := services.InventoryResult{
		Success:     true,
		Message:     "Inventory created successfully",
		InventoryID: inventoryID,
		ProductID:   productID,
		Quantity:    100,
	}

	// Test the fields
	assert.True(t, result.Success)
	assert.Equal(t, "Inventory created successfully", result.Message)
	assert.Equal(t, inventoryID, result.InventoryID)
	assert.Equal(t, productID, result.ProductID)
	assert.Equal(t, 100, result.Quantity)
}

// TestPriceResult tests the PriceResult struct
func TestPriceResult(t *testing.T) {
	// Create a PriceResult
	priceID := uuid.New()
	productID := uuid.New()
	result := services.PriceResult{
		Success:   true,
		Message:   "Price created successfully",
		PriceID:   priceID,
		ProductID: productID,
		Price:     1000.0,
		Currency:  "VND",
	}

	// Test the fields
	assert.True(t, result.Success)
	assert.Equal(t, "Price created successfully", result.Message)
	assert.Equal(t, priceID, result.PriceID)
	assert.Equal(t, productID, result.ProductID)
	assert.Equal(t, 1000.0, result.Price)
	assert.Equal(t, "VND", result.Currency)
}

// TestProduct tests the Product model
func TestProduct(t *testing.T) {
	// Create a Product
	productID := uuid.New()
	p := product.Product{
		Name:        "Test Product",
		SKU:         "TEST-123",
		Description: "This is a test product",
		Category:    "Test Category",
		ImageURL:    "http://example.com/image.jpg",
	}
	p.ID = productID

	// Test the fields
	assert.Equal(t, productID, p.ID)
	assert.Equal(t, "Test Product", p.Name)
	assert.Equal(t, "TEST-123", p.SKU)
	assert.Equal(t, "This is a test product", p.Description)
	assert.Equal(t, "Test Category", p.Category)
	assert.Equal(t, "http://example.com/image.jpg", p.ImageURL)
}

// TestInventory tests the Inventory model
func TestInventory(t *testing.T) {
	// Create an Inventory
	inventoryID := uuid.New()
	productID := uuid.New()
	inv := product.Inventory{
		ProductID: productID,
		Size:      "M",
		Color:     "Blue",
		Quantity:  100,
		Location:  "Warehouse A",
	}
	inv.ID = inventoryID

	// Test the fields
	assert.Equal(t, inventoryID, inv.ID)
	assert.Equal(t, productID, inv.ProductID)
	assert.Equal(t, "M", inv.Size)
	assert.Equal(t, "Blue", inv.Color)
	assert.Equal(t, 100, inv.Quantity)
	assert.Equal(t, "Warehouse A", inv.Location)
}

// TestPrice tests the Price model
func TestPrice(t *testing.T) {
	// Create a Price
	priceID := uuid.New()
	productID := uuid.New()
	now := time.Now()
	endDate := now.Add(30 * 24 * time.Hour) // 30 days later
	price := product.Price{
		ProductID: productID,
		Price:     1000.0,
		Currency:  "VND",
		StartDate: now,
		EndDate:   &endDate,
	}
	price.ID = priceID

	// Test the fields
	assert.Equal(t, priceID, price.ID)
	assert.Equal(t, productID, price.ProductID)
	assert.Equal(t, 1000.0, price.Price)
	assert.Equal(t, "VND", price.Currency)
	assert.Equal(t, now.Unix(), price.StartDate.Unix())
	assert.Equal(t, endDate.Unix(), price.EndDate.Unix())
}
