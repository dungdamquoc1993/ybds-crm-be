package responses

import (
	"time"

	"github.com/google/uuid"
)

// AddressResponse defines the address data in the response
type AddressResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	GuestID   *uuid.UUID `json:"guest_id,omitempty"`
	Address   string     `json:"address"`
	Ward      string     `json:"ward"`
	District  string     `json:"district"`
	City      string     `json:"city"`
	Country   string     `json:"country"`
	IsDefault bool       `json:"is_default"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// UserDetailResponse defines the detailed user data in the response
type UserDetailResponse struct {
	ID        uuid.UUID         `json:"id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Phone     string            `json:"phone"`
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name"`
	IsActive  bool              `json:"is_active"`
	Roles     []string          `json:"roles"`
	Addresses []AddressResponse `json:"addresses"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// UsersResponse defines the response for a list of users
type UsersResponse struct {
	Success    bool                 `json:"success"`
	Message    string               `json:"message"`
	Users      []UserDetailResponse `json:"users"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// GuestDetailResponse defines the detailed guest data in the response
type GuestDetailResponse struct {
	ID        uuid.UUID         `json:"id"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	Phone     string            `json:"phone"`
	Addresses []AddressResponse `json:"addresses"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
