package responses

import (
	"time"

	"github.com/google/uuid"
	"github.com/ybds/internal/models/product"
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
	Inventories []InventoryResponse `json:"inventories,omitempty"`
	Prices      []PriceResponse     `json:"prices,omitempty"`
	Images      []ImageResponse     `json:"images,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// ImageResponse defines the image data in a response
type ImageResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	IsPrimary bool      `json:"is_primary"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

// ConvertToInventoryResponse converts a product.Inventory to an InventoryResponse
func ConvertToInventoryResponse(inventory product.Inventory) InventoryResponse {
	return InventoryResponse{
		ID:        inventory.ID,
		ProductID: inventory.ProductID,
		Size:      inventory.Size,
		Color:     inventory.Color,
		Quantity:  inventory.Quantity,
		Location:  inventory.Location,
		CreatedAt: inventory.CreatedAt,
		UpdatedAt: inventory.UpdatedAt,
	}
}

// ConvertToPriceResponse converts a product.Price to a PriceResponse
func ConvertToPriceResponse(price product.Price) PriceResponse {
	return PriceResponse{
		ID:        price.ID,
		ProductID: price.ProductID,
		Price:     price.Price,
		Currency:  price.Currency,
		StartDate: price.StartDate,
		EndDate:   price.EndDate,
		CreatedAt: price.CreatedAt,
		UpdatedAt: price.UpdatedAt,
	}
}

// ConvertToImageResponse converts a product.ProductImage to an ImageResponse
func ConvertToImageResponse(image product.ProductImage) ImageResponse {
	return ImageResponse{
		ID:        image.ID,
		ProductID: image.ProductID,
		URL:       image.URL,
		Filename:  image.Filename,
		IsPrimary: image.IsPrimary,
		SortOrder: image.SortOrder,
		CreatedAt: image.CreatedAt,
		UpdatedAt: image.UpdatedAt,
	}
}

// ConvertToProductDetailResponse converts a product.Product to a ProductDetailResponse
func ConvertToProductDetailResponse(p product.Product) ProductDetailResponse {
	response := ProductDetailResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		SKU:         p.SKU,
		Category:    p.Category,
		ImageURL:    p.ImageURL,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	// Convert inventories
	if len(p.Inventory) > 0 {
		response.Inventories = make([]InventoryResponse, len(p.Inventory))
		for i, inv := range p.Inventory {
			response.Inventories[i] = ConvertToInventoryResponse(inv)
		}
	}

	// Convert prices
	if len(p.Prices) > 0 {
		response.Prices = make([]PriceResponse, len(p.Prices))
		for i, price := range p.Prices {
			response.Prices[i] = ConvertToPriceResponse(price)
		}
	}

	// Convert images
	if len(p.Images) > 0 {
		response.Images = make([]ImageResponse, len(p.Images))
		for i, img := range p.Images {
			response.Images[i] = ConvertToImageResponse(img)
		}
	}

	return response
}

// ConvertToProductDetailResponses converts a slice of product.Product to a slice of ProductDetailResponse
func ConvertToProductDetailResponses(products []product.Product) []ProductDetailResponse {
	responses := make([]ProductDetailResponse, len(products))
	for i, p := range products {
		responses[i] = ConvertToProductDetailResponse(p)
	}
	return responses
}
