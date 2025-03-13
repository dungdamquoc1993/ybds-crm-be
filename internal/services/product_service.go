package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ybds/internal/models/product"
	"github.com/ybds/internal/repositories"
	"gorm.io/gorm"
)

// ProductService handles product-related business logic
type ProductService struct {
	DB                  *gorm.DB
	ProductRepo         *repositories.ProductRepository
	NotificationService *NotificationService
}

// NewProductService creates a new instance of ProductService
func NewProductService(db *gorm.DB, notificationService *NotificationService) *ProductService {
	return &ProductService{
		DB:                  db,
		ProductRepo:         repositories.NewProductRepository(db),
		NotificationService: notificationService,
	}
}

// ProductResult represents the result of a product operation
type ProductResult struct {
	Success   bool
	Message   string
	Error     string
	ProductID uuid.UUID
	Name      string
	SKU       string
}

// GetProductByID retrieves a product by ID
func (s *ProductService) GetProductByID(id uuid.UUID) (*product.Product, error) {
	return s.ProductRepo.GetProductByID(id)
}

// GetProductBySKU retrieves a product by SKU
func (s *ProductService) GetProductBySKU(sku string) (*product.Product, error) {
	return s.ProductRepo.GetProductBySKU(sku)
}

// GetAllProducts retrieves all products with pagination and filtering
func (s *ProductService) GetAllProducts(page, pageSize int, filters map[string]interface{}) ([]product.Product, int64, error) {
	return s.ProductRepo.GetAllProducts(page, pageSize, filters)
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(name, description, sku, category, imageURL string) (*ProductResult, error) {
	// Validate input
	if name == "" {
		return &ProductResult{
			Success: false,
			Message: "Product creation failed",
			Error:   "Name is required",
		}, fmt.Errorf("name is required")
	}

	if sku == "" {
		return &ProductResult{
			Success: false,
			Message: "Product creation failed",
			Error:   "SKU is required",
		}, fmt.Errorf("sku is required")
	}

	if category == "" {
		return &ProductResult{
			Success: false,
			Message: "Product creation failed",
			Error:   "Category is required",
		}, fmt.Errorf("category is required")
	}

	// Check if product with SKU already exists
	existingProduct, err := s.ProductRepo.GetProductBySKU(sku)
	if err == nil && existingProduct != nil && existingProduct.ID != uuid.Nil {
		return &ProductResult{
			Success: false,
			Message: "Product creation failed",
			Error:   "Product with this SKU already exists",
		}, fmt.Errorf("product with SKU %s already exists", sku)
	}

	// Create product
	p := &product.Product{
		Name:        name,
		Description: description,
		SKU:         sku,
		Category:    category,
		ImageURL:    imageURL,
	}

	// Save product
	if err := s.ProductRepo.CreateProduct(p); err != nil {
		return &ProductResult{
			Success: false,
			Message: "Product creation failed",
			Error:   "Error creating product",
		}, err
	}

	// Send notification
	if s.NotificationService != nil {
		metadata := map[string]interface{}{
			"product_id":   p.ID.String(),
			"product_name": p.Name,
			"sku":          p.SKU,
			"category":     p.Category,
		}
		s.NotificationService.CreateProductNotification(p.ID, p.Name, "created", metadata)
	}

	return &ProductResult{
		Success:   true,
		Message:   "Product created successfully",
		ProductID: p.ID,
		Name:      p.Name,
		SKU:       p.SKU,
	}, nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(id uuid.UUID, name, description, sku, category, imageURL string) (*ProductResult, error) {
	// Get the product
	p, err := s.ProductRepo.GetProductByID(id)
	if err != nil {
		return &ProductResult{
			Success: false,
			Message: "Product update failed",
			Error:   "Product not found",
		}, err
	}

	// Update fields if provided
	if name != "" {
		p.Name = name
	}
	if description != "" {
		p.Description = description
	}
	if sku != "" && sku != p.SKU {
		// Check if product with new SKU already exists
		existingProduct, err := s.ProductRepo.GetProductBySKU(sku)
		if err == nil && existingProduct != nil && existingProduct.ID != uuid.Nil && existingProduct.ID != id {
			return &ProductResult{
				Success: false,
				Message: "Product update failed",
				Error:   "Product with this SKU already exists",
			}, fmt.Errorf("product with SKU %s already exists", sku)
		}
		p.SKU = sku
	}
	if category != "" {
		p.Category = category
	}
	if imageURL != "" {
		p.ImageURL = imageURL
	}

	// Save product
	if err := s.ProductRepo.UpdateProduct(p); err != nil {
		return &ProductResult{
			Success: false,
			Message: "Product update failed",
			Error:   "Error updating product",
		}, err
	}

	// Send notification
	if s.NotificationService != nil {
		metadata := map[string]interface{}{
			"product_id":   p.ID.String(),
			"product_name": p.Name,
			"sku":          p.SKU,
			"category":     p.Category,
		}
		s.NotificationService.CreateProductNotification(p.ID, p.Name, "updated", metadata)
	}

	return &ProductResult{
		Success:   true,
		Message:   "Product updated successfully",
		ProductID: p.ID,
		Name:      p.Name,
		SKU:       p.SKU,
	}, nil
}

// DeleteProduct deletes a product by ID
func (s *ProductService) DeleteProduct(id uuid.UUID) (*ProductResult, error) {
	// Get the product
	p, err := s.ProductRepo.GetProductByID(id)
	if err != nil {
		return &ProductResult{
			Success: false,
			Message: "Product deletion failed",
			Error:   "Product not found",
		}, err
	}

	// Delete the product
	if err := s.ProductRepo.DeleteProduct(id); err != nil {
		return &ProductResult{
			Success: false,
			Message: "Product deletion failed",
			Error:   "Error deleting product",
		}, err
	}

	// Send notification
	if s.NotificationService != nil {
		metadata := map[string]interface{}{
			"product_id":   p.ID.String(),
			"product_name": p.Name,
			"sku":          p.SKU,
			"category":     p.Category,
		}
		s.NotificationService.CreateProductNotification(p.ID, p.Name, "deleted", metadata)
	}

	return &ProductResult{
		Success:   true,
		Message:   "Product deleted successfully",
		ProductID: p.ID,
		Name:      p.Name,
		SKU:       p.SKU,
	}, nil
}

// InventoryResult represents the result of an inventory operation
type InventoryResult struct {
	Success     bool
	Message     string
	Error       string
	InventoryID uuid.UUID
	ProductID   uuid.UUID
	Quantity    int
}

// GetInventoryByID retrieves an inventory by ID
func (s *ProductService) GetInventoryByID(id uuid.UUID) (*product.Inventory, error) {
	return s.ProductRepo.GetInventoryByID(id)
}

// GetInventoriesByProductID retrieves all inventories for a product
func (s *ProductService) GetInventoriesByProductID(productID uuid.UUID) ([]product.Inventory, error) {
	return s.ProductRepo.GetInventoriesByProductID(productID)
}

// CheckInventoryAvailability checks if there is enough inventory for the given quantity
func (s *ProductService) CheckInventoryAvailability(inventoryID uuid.UUID, quantity int) (bool, error) {
	inventory, err := s.ProductRepo.GetInventoryByID(inventoryID)
	if err != nil {
		return false, err
	}
	return inventory.Quantity >= quantity, nil
}

// CreateInventory creates a new inventory
func (s *ProductService) CreateInventory(productID uuid.UUID, size, color string, quantity int, location string) (*InventoryResult, error) {
	// Validate input
	if productID == uuid.Nil {
		return &InventoryResult{
			Success: false,
			Message: "Inventory creation failed",
			Error:   "Product ID is required",
		}, fmt.Errorf("product ID is required")
	}

	// Check if product exists
	p, err := s.ProductRepo.GetProductByID(productID)
	if err != nil {
		return &InventoryResult{
			Success: false,
			Message: "Inventory creation failed",
			Error:   "Product not found",
		}, err
	}

	// Create inventory
	inventory := &product.Inventory{
		ProductID: productID,
		Size:      size,
		Color:     color,
		Quantity:  quantity,
		Location:  location,
	}

	// Save inventory
	if err := s.ProductRepo.CreateInventory(inventory); err != nil {
		return &InventoryResult{
			Success: false,
			Message: "Inventory creation failed",
			Error:   "Error creating inventory",
		}, err
	}

	// Send notification if quantity is low
	if s.NotificationService != nil && quantity <= 5 {
		metadata := map[string]interface{}{
			"product_id":   p.ID.String(),
			"product_name": p.Name,
			"inventory_id": inventory.ID.String(),
			"quantity":     quantity,
			"size":         size,
			"color":        color,
		}

		event := "low_stock"
		if quantity == 0 {
			event = "out_of_stock"
		}

		s.NotificationService.CreateProductNotification(p.ID, p.Name, event, metadata)
	}

	return &InventoryResult{
		Success:     true,
		Message:     "Inventory created successfully",
		InventoryID: inventory.ID,
		ProductID:   productID,
		Quantity:    quantity,
	}, nil
}

// UpdateInventory updates an existing inventory
func (s *ProductService) UpdateInventory(id uuid.UUID, size, color string, quantity *int, location string) (*InventoryResult, error) {
	// Get the inventory
	inventory, err := s.ProductRepo.GetInventoryByID(id)
	if err != nil {
		return &InventoryResult{
			Success: false,
			Message: "Inventory update failed",
			Error:   "Inventory not found",
		}, err
	}

	// Get the product
	p, err := s.ProductRepo.GetProductByID(inventory.ProductID)
	if err != nil {
		return &InventoryResult{
			Success: false,
			Message: "Inventory update failed",
			Error:   "Product not found",
		}, err
	}

	// Update fields if provided
	if size != "" {
		inventory.Size = size
	}
	if color != "" {
		inventory.Color = color
	}

	oldQuantity := inventory.Quantity

	if quantity != nil {
		inventory.Quantity = *quantity
	}
	if location != "" {
		inventory.Location = location
	}

	// Save inventory
	if err := s.ProductRepo.UpdateInventory(inventory); err != nil {
		return &InventoryResult{
			Success: false,
			Message: "Inventory update failed",
			Error:   "Error updating inventory",
		}, err
	}

	// Send notification if quantity changed to low or zero
	if s.NotificationService != nil && quantity != nil {
		// Check if quantity changed significantly
		if (oldQuantity > 5 && *quantity <= 5) || (oldQuantity > 0 && *quantity == 0) {
			metadata := map[string]interface{}{
				"product_id":   p.ID.String(),
				"product_name": p.Name,
				"inventory_id": inventory.ID.String(),
				"quantity":     *quantity,
				"size":         inventory.Size,
				"color":        inventory.Color,
			}

			event := "low_stock"
			if *quantity == 0 {
				event = "out_of_stock"
			}

			s.NotificationService.CreateProductNotification(p.ID, p.Name, event, metadata)
		} else if oldQuantity == 0 && *quantity > 0 {
			// Back in stock notification
			metadata := map[string]interface{}{
				"product_id":   p.ID.String(),
				"product_name": p.Name,
				"inventory_id": inventory.ID.String(),
				"quantity":     *quantity,
				"size":         inventory.Size,
				"color":        inventory.Color,
			}

			s.NotificationService.CreateProductNotification(p.ID, p.Name, "back_in_stock", metadata)
		}
	}

	return &InventoryResult{
		Success:     true,
		Message:     "Inventory updated successfully",
		InventoryID: inventory.ID,
		ProductID:   inventory.ProductID,
		Quantity:    inventory.Quantity,
	}, nil
}

// DeleteInventory deletes an inventory by ID
func (s *ProductService) DeleteInventory(id uuid.UUID) (*InventoryResult, error) {
	// Get the inventory
	inventory, err := s.ProductRepo.GetInventoryByID(id)
	if err != nil {
		return &InventoryResult{
			Success: false,
			Message: "Inventory deletion failed",
			Error:   "Inventory not found",
		}, err
	}

	// Delete the inventory
	if err := s.ProductRepo.DeleteInventory(id); err != nil {
		return &InventoryResult{
			Success: false,
			Message: "Inventory deletion failed",
			Error:   "Error deleting inventory",
		}, err
	}

	return &InventoryResult{
		Success:     true,
		Message:     "Inventory deleted successfully",
		InventoryID: id,
		ProductID:   inventory.ProductID,
		Quantity:    inventory.Quantity,
	}, nil
}

// PriceResult represents the result of a price operation
type PriceResult struct {
	Success   bool
	Message   string
	Error     string
	PriceID   uuid.UUID
	ProductID uuid.UUID
	Price     float64
	Currency  string
}

// GetPriceByID retrieves a price by ID
func (s *ProductService) GetPriceByID(id uuid.UUID) (*product.Price, error) {
	return s.ProductRepo.GetPriceByID(id)
}

// GetPricesByProductID retrieves all prices for a product
func (s *ProductService) GetPricesByProductID(productID uuid.UUID) ([]product.Price, error) {
	return s.ProductRepo.GetPricesByProductID(productID)
}

// GetCurrentPrice retrieves the current valid price for a product
func (s *ProductService) GetCurrentPrice(productID uuid.UUID) (*product.Price, error) {
	return s.ProductRepo.GetCurrentPrice(productID)
}

// CreatePrice creates a new price
func (s *ProductService) CreatePrice(productID uuid.UUID, price float64, currency string, startDate time.Time, endDate *time.Time) (*PriceResult, error) {
	// Validate input
	if productID == uuid.Nil {
		return &PriceResult{
			Success: false,
			Message: "Price creation failed",
			Error:   "Product ID is required",
		}, fmt.Errorf("product ID is required")
	}

	if price <= 0 {
		return &PriceResult{
			Success: false,
			Message: "Price creation failed",
			Error:   "Price must be greater than zero",
		}, fmt.Errorf("price must be greater than zero")
	}

	// Check if product exists
	_, err := s.ProductRepo.GetProductByID(productID)
	if err != nil {
		return &PriceResult{
			Success: false,
			Message: "Price creation failed",
			Error:   "Product not found",
		}, err
	}

	// Create price
	p := &product.Price{
		ProductID: productID,
		Price:     price,
		Currency:  currency,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Save price
	if err := s.ProductRepo.CreatePrice(p); err != nil {
		return &PriceResult{
			Success: false,
			Message: "Price creation failed",
			Error:   "Error creating price",
		}, err
	}

	return &PriceResult{
		Success:   true,
		Message:   "Price created successfully",
		PriceID:   p.ID,
		ProductID: productID,
		Price:     price,
		Currency:  currency,
	}, nil
}

// UpdatePrice updates an existing price
func (s *ProductService) UpdatePrice(id uuid.UUID, price *float64, currency string, startDate *time.Time, endDate *time.Time) (*PriceResult, error) {
	// Get the price
	p, err := s.ProductRepo.GetPriceByID(id)
	if err != nil {
		return &PriceResult{
			Success: false,
			Message: "Price update failed",
			Error:   "Price not found",
		}, err
	}

	// Update fields if provided
	if price != nil {
		if *price <= 0 {
			return &PriceResult{
				Success: false,
				Message: "Price update failed",
				Error:   "Price must be greater than zero",
			}, fmt.Errorf("price must be greater than zero")
		}
		p.Price = *price
	}
	if currency != "" {
		p.Currency = currency
	}
	if startDate != nil {
		p.StartDate = *startDate
	}
	if endDate != nil {
		p.EndDate = endDate
	}

	// Save price
	if err := s.ProductRepo.UpdatePrice(p); err != nil {
		return &PriceResult{
			Success: false,
			Message: "Price update failed",
			Error:   "Error updating price",
		}, err
	}

	return &PriceResult{
		Success:   true,
		Message:   "Price updated successfully",
		PriceID:   p.ID,
		ProductID: p.ProductID,
		Price:     p.Price,
		Currency:  p.Currency,
	}, nil
}

// DeletePrice deletes a price by ID
func (s *ProductService) DeletePrice(id uuid.UUID) (*PriceResult, error) {
	// Get the price
	p, err := s.ProductRepo.GetPriceByID(id)
	if err != nil {
		return &PriceResult{
			Success: false,
			Message: "Price deletion failed",
			Error:   "Price not found",
		}, err
	}

	// Delete the price
	if err := s.ProductRepo.DeletePrice(id); err != nil {
		return &PriceResult{
			Success: false,
			Message: "Price deletion failed",
			Error:   "Error deleting price",
		}, err
	}

	return &PriceResult{
		Success:   true,
		Message:   "Price deleted successfully",
		PriceID:   id,
		ProductID: p.ProductID,
		Price:     p.Price,
		Currency:  p.Currency,
	}, nil
}

// ReserveInventory reduces the inventory quantity by the given amount
func (s *ProductService) ReserveInventory(inventoryID uuid.UUID, quantity int) error {
	inventory, err := s.ProductRepo.GetInventoryByID(inventoryID)
	if err != nil {
		return err
	}

	if inventory.Quantity < quantity {
		return fmt.Errorf("not enough inventory")
	}

	inventory.Quantity -= quantity
	return s.ProductRepo.UpdateInventory(inventory)
}

// ReleaseInventory increases the inventory quantity by the given amount
func (s *ProductService) ReleaseInventory(inventoryID uuid.UUID, quantity int) error {
	inventory, err := s.ProductRepo.GetInventoryByID(inventoryID)
	if err != nil {
		return err
	}

	inventory.Quantity += quantity
	return s.ProductRepo.UpdateInventory(inventory)
}
