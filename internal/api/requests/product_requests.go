package requests

import (
	"fmt"
	"strings"
	"time"
)

// InventoryRequest defines the inventory data in a request
type InventoryRequest struct {
	Size     string `json:"size"`
	Color    string `json:"color"`
	Quantity int    `json:"quantity"`
	Location string `json:"location"`
}

// PriceRequest defines the price data in a request
type PriceRequest struct {
	Price    float64    `json:"price"`
	Currency string     `json:"currency"`
	EndDate  *time.Time `json:"end_date,omitempty"`
}

// CreateProductRequest defines the request model for creating a product
type CreateProductRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	SKU         string             `json:"sku"`
	Category    string             `json:"category"`
	ImageURL    string             `json:"image_url"`
	Inventories []InventoryRequest `json:"inventories,omitempty"`
	Prices      []PriceRequest     `json:"prices,omitempty"`
}

// Validate validates the create product request
func (r *CreateProductRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.SKU = strings.TrimSpace(r.SKU)
	r.Category = strings.TrimSpace(r.Category)

	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if r.SKU == "" {
		return fmt.Errorf("SKU is required")
	}

	if r.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate inventories if provided
	for i, inv := range r.Inventories {
		if inv.Quantity < 0 {
			return fmt.Errorf("inventory %d: quantity cannot be negative", i+1)
		}
	}

	// Validate prices if provided
	for i, price := range r.Prices {
		if price.Price <= 0 {
			return fmt.Errorf("price %d: price must be greater than zero", i+1)
		}

		if price.Currency == "" {
			return fmt.Errorf("price %d: currency is required", i+1)
		}
	}

	return nil
}

// UpdateProductRequest defines the request model for updating a product
type UpdateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SKU         string `json:"sku"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
}

// CreateInventoryRequest defines the request model for creating an inventory
type CreateInventoryRequest struct {
	Size     string `json:"size"`
	Color    string `json:"color"`
	Quantity int    `json:"quantity"`
	Location string `json:"location"`
}

// Validate validates the create inventory request
func (r *CreateInventoryRequest) Validate() error {
	if r.Quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}

	return nil
}

// CreateMultipleInventoriesRequest defines the request model for creating multiple inventories
type CreateMultipleInventoriesRequest struct {
	Inventories []CreateInventoryRequest `json:"inventories"`
}

// Validate validates the create multiple inventories request
func (r *CreateMultipleInventoriesRequest) Validate() error {
	if len(r.Inventories) == 0 {
		return fmt.Errorf("at least one inventory is required")
	}

	for i, inv := range r.Inventories {
		if err := inv.Validate(); err != nil {
			return fmt.Errorf("inventory %d: %s", i+1, err.Error())
		}
	}

	return nil
}

// UpdateInventoryRequest defines the request model for updating an inventory
type UpdateInventoryRequest struct {
	Size     string `json:"size"`
	Color    string `json:"color"`
	Quantity int    `json:"quantity"`
	Location string `json:"location"`
}

// CreatePriceRequest defines the request model for creating a price
type CreatePriceRequest struct {
	Price    float64    `json:"price"`
	Currency string     `json:"currency"`
	EndDate  *time.Time `json:"end_date,omitempty"`
}

// Validate validates the create price request
func (r *CreatePriceRequest) Validate() error {
	if r.Price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}

	r.Currency = strings.TrimSpace(r.Currency)
	if r.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	return nil
}

// UpdatePriceRequest defines the request model for updating a price
type UpdatePriceRequest struct {
	Price    float64    `json:"price"`
	Currency string     `json:"currency"`
	EndDate  *time.Time `json:"end_date,omitempty"`
}
