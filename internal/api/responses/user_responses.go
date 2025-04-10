package responses

import (
	"time"

	"github.com/google/uuid"
)

// UserDetailResponse defines the detailed user data in the response
type UserDetailResponse struct {
	ID         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	IsActive   bool      `json:"is_active"`
	TelegramID int64     `json:"telegram_id"`
	Roles      []string  `json:"roles"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SingleUserResponse defines the response for a single user
type SingleUserResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    UserDetailResponse `json:"data"`
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
