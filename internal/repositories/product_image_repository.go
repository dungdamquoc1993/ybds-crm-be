package repositories

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models/product"
	"gorm.io/gorm"
)

// ProductImageRepository handles database operations for product images
type ProductImageRepository struct {
	db *gorm.DB
}

// NewProductImageRepository creates a new instance of ProductImageRepository
func NewProductImageRepository(db *gorm.DB) *ProductImageRepository {
	return &ProductImageRepository{
		db: db,
	}
}

// GetImagesByProductID retrieves all images for a product
func (r *ProductImageRepository) GetImagesByProductID(productID uuid.UUID) ([]product.ProductImage, error) {
	// Check if product exists and is not deleted
	var count int64
	if err := r.db.Model(&product.Product{}).Where("id = ? AND deleted_at IS NULL", productID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var images []product.ProductImage
	err := r.db.Where("product_id = ?", productID).
		Order("is_primary DESC, sort_order ASC").
		Find(&images).Error
	return images, err
}

// GetImageByID retrieves an image by ID
func (r *ProductImageRepository) GetImageByID(id uuid.UUID) (*product.ProductImage, error) {
	var image product.ProductImage
	err := r.db.Joins("JOIN products ON product_images.product_id = products.id").
		Where("product_images.id = ? AND products.deleted_at IS NULL", id).
		First(&image).Error
	return &image, err
}

// CreateImage creates a new product image
func (r *ProductImageRepository) CreateImage(image *product.ProductImage) error {
	return r.db.Create(image).Error
}

// UpdateImage updates an existing product image
func (r *ProductImageRepository) UpdateImage(image *product.ProductImage) error {
	return r.db.Save(image).Error
}

// DeleteImage deletes a product image by ID
func (r *ProductImageRepository) DeleteImage(id uuid.UUID) error {
	return r.db.Delete(&product.ProductImage{}, id).Error
}

// SetPrimaryImage sets an image as the primary image for a product
func (r *ProductImageRepository) SetPrimaryImage(imageID, productID uuid.UUID) error {
	// Start a transaction
	tx := r.db.Begin()

	// Reset all images for this product to not primary
	if err := tx.Model(&product.ProductImage{}).
		Where("product_id = ?", productID).
		Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set the specified image as primary
	if err := tx.Model(&product.ProductImage{}).
		Where("id = ? AND product_id = ?", imageID, productID).
		Update("is_primary", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

// ReorderImages updates the sort order of images
func (r *ProductImageRepository) ReorderImages(productID uuid.UUID, imageIDs []uuid.UUID) error {
	// Start a transaction
	tx := r.db.Begin()

	// Update the sort order for each image
	for i, imageID := range imageIDs {
		if err := tx.Model(&product.ProductImage{}).
			Where("id = ? AND product_id = ?", imageID, productID).
			Update("sort_order", i).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	return tx.Commit().Error
}
