package responses

import (
	"time"

	"github.com/google/uuid"
)

// InventoryResponse defines the inventory data in a response
type InventoryResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Size      string    `json:"size"`
	Color     string    `json:"color"`
	Quantity  int       `json:"quantity"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PriceResponse defines the price data in a response
type PriceResponse struct {
	ID        uuid.UUID  `json:"id"`
	ProductID uuid.UUID  `json:"product_id"`
	Price     float64    `json:"price"`
	Currency  string     `json:"currency"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ProductResponse defines the product data in a response
type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SKU         string    `json:"sku"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductDetailResponse defines the detailed product data in a response
type ProductDetailResponse struct {
	ID          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	SKU         string              `json:"sku"`
	Category    string              `json:"category"`
	ImageURL    string              `json:"image_url"`
	Inventories []InventoryResponse `json:"inventories"`
	Prices      []PriceResponse     `json:"prices"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// ProductsResponse defines the response for a list of products
type ProductsResponse struct {
	Success    bool                    `json:"success"`
	Message    string                  `json:"message"`
	Products   []ProductDetailResponse `json:"products"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

// SuccessResponse defines a standard success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
