package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/ybds/internal/models/product"
	"gorm.io/gorm"
)

// ProductRepository handles database operations for products
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new instance of ProductRepository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// GetProductByID retrieves a product by ID with all relations
func (r *ProductRepository) GetProductByID(id uuid.UUID) (*product.Product, error) {
	var p product.Product
	err := r.db.Where("id = ?", id).
		Preload("Inventory").
		Preload("Prices").
		Preload("Images").
		First(&p).Error
	return &p, err
}

// GetProductBySKU retrieves a product by SKU with all relations
func (r *ProductRepository) GetProductBySKU(sku string) (*product.Product, error) {
	var p product.Product
	err := r.db.Where("sku = ?", sku).
		Preload("Inventory").
		Preload("Prices").
		Preload("Images").
		First(&p).Error
	return &p, err
}

// GetAllProducts retrieves all products with pagination and filtering
func (r *ProductRepository) GetAllProducts(page, pageSize int, filters map[string]interface{}) ([]product.Product, int64, error) {
	var products []product.Product
	var total int64

	query := r.db.Model(&product.Product{})

	// Apply filters
	for key, value := range filters {
		switch key {
		case "name":
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		case "category":
			query = query.Where("category = ?", value)
		case "sku":
			query = query.Where("sku LIKE ?", "%"+value.(string)+"%")
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Preload("Inventory").
		Preload("Prices").
		Preload("Images").
		Find(&products).Error

	return products, total, err
}

// CreateProduct creates a new product
func (r *ProductRepository) CreateProduct(p *product.Product) error {
	return r.db.Create(p).Error
}

// UpdateProduct updates an existing product
func (r *ProductRepository) UpdateProduct(p *product.Product) error {
	return r.db.Save(p).Error
}

// DeleteProduct deletes a product by ID
func (r *ProductRepository) DeleteProduct(id uuid.UUID) error {
	return r.db.Delete(&product.Product{}, id).Error
}

// GetInventoryByID retrieves an inventory by ID
func (r *ProductRepository) GetInventoryByID(id uuid.UUID) (*product.Inventory, error) {
	var inventory product.Inventory
	// Join with products table and check if product is not deleted
	err := r.db.Joins("JOIN products ON inventory.product_id = products.id").
		Where("inventory.id = ? AND products.deleted_at IS NULL", id).
		First(&inventory).Error
	return &inventory, err
}

// GetInventoriesByProductID retrieves all inventories for a product
func (r *ProductRepository) GetInventoriesByProductID(productID uuid.UUID) ([]product.Inventory, error) {
	var inventories []product.Inventory
	// Check if product exists and is not deleted
	var count int64
	if err := r.db.Model(&product.Product{}).Where("id = ? AND deleted_at IS NULL", productID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Get inventories for the product
	err := r.db.Where("product_id = ?", productID).Find(&inventories).Error
	return inventories, err
}

// CreateInventory creates a new inventory
func (r *ProductRepository) CreateInventory(inventory *product.Inventory) error {
	return r.db.Create(inventory).Error
}

// UpdateInventory updates an existing inventory
func (r *ProductRepository) UpdateInventory(inventory *product.Inventory) error {
	return r.db.Save(inventory).Error
}

// DeleteInventory deletes an inventory by ID
func (r *ProductRepository) DeleteInventory(id uuid.UUID) error {
	return r.db.Delete(&product.Inventory{}, id).Error
}

// GetPriceByID retrieves a price by ID
func (r *ProductRepository) GetPriceByID(id uuid.UUID) (*product.Price, error) {
	var price product.Price
	err := r.db.Joins("JOIN products ON prices.product_id = products.id").
		Where("prices.id = ? AND products.deleted_at IS NULL", id).
		First(&price).Error
	return &price, err
}

// GetPricesByProductID retrieves all prices for a product
func (r *ProductRepository) GetPricesByProductID(productID uuid.UUID) ([]product.Price, error) {
	var prices []product.Price

	// Check if product exists and is not deleted
	var count int64
	if err := r.db.Model(&product.Product{}).Where("id = ? AND deleted_at IS NULL", productID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	err := r.db.Where("product_id = ?", productID).Find(&prices).Error
	return prices, err
}

// GetCurrentPrice retrieves the current valid price for an inventory
func (r *ProductRepository) GetCurrentPrice(productID uuid.UUID) (*product.Price, error) {
	var price product.Price
	now := time.Now()

	// Check if product exists and is not deleted
	var count int64
	if err := r.db.Model(&product.Product{}).Where("id = ? AND deleted_at IS NULL", productID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	err := r.db.Where("product_id = ? AND start_date <= ? AND (end_date IS NULL OR end_date > ?)",
		productID, now, now).
		Order("start_date DESC, created_at DESC").
		First(&price).Error

	return &price, err
}

// CreatePrice creates a new price
func (r *ProductRepository) CreatePrice(price *product.Price) error {
	return r.db.Create(price).Error
}

// UpdatePrice updates an existing price
func (r *ProductRepository) UpdatePrice(price *product.Price) error {
	return r.db.Save(price).Error
}

// DeletePrice deletes a price by ID
func (r *ProductRepository) DeletePrice(id uuid.UUID) error {
	return r.db.Delete(&product.Price{}, id).Error
}

// CreateInventoryTransaction creates a new inventory transaction
func (r *ProductRepository) CreateInventoryTransaction(transaction *product.InventoryTransaction) error {
	return r.db.Create(transaction).Error
}

// GetInventoryTransactionsByInventoryID retrieves all transactions for an inventory
func (r *ProductRepository) GetInventoryTransactionsByInventoryID(inventoryID uuid.UUID) ([]product.InventoryTransaction, error) {
	var transactions []product.InventoryTransaction
	err := r.db.Where("inventory_id = ?", inventoryID).Find(&transactions).Error
	return transactions, err
}

// UpdateInventoryQuantity updates the quantity of an inventory and creates a transaction
func (r *ProductRepository) UpdateInventoryQuantity(inventoryID uuid.UUID, quantity int, txType product.TransactionType, reason product.TransactionReason, referenceID *uuid.UUID, referenceType string, notes string) error {
	// Start a transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Get the inventory
	var inventory product.Inventory
	if err := tx.Where("id = ?", inventoryID).First(&inventory).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update the inventory quantity
	inventory.Quantity += quantity
	if err := tx.Save(&inventory).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create a transaction record
	transaction := product.InventoryTransaction{
		InventoryID:   inventoryID,
		Quantity:      quantity,
		Type:          txType,
		Reason:        reason,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		Notes:         notes,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}
