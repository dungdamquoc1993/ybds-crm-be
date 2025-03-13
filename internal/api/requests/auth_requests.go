package requests

import (
	"fmt"
	"strings"
)

// LoginRequest defines the login request model
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)

	if r.Username == "" {
		return fmt.Errorf("username is required")
	}

	if r.Password == "" {
		return fmt.Errorf("password is required")
	}

	if len(r.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	return nil
}

// RegisterRequest defines the registration request model
type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// Validate validates the registration request
func (r *RegisterRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	r.Phone = strings.TrimSpace(r.Phone)
	r.Password = strings.TrimSpace(r.Password)

	// Either email or phone is required
	if r.Email == "" && r.Phone == "" {
		return fmt.Errorf("email or phone number is required")
	}

	// Validate password
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}

	if len(r.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	return nil
}
