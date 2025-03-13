package requests

import (
	"errors"
	"regexp"
	"strings"
)

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	Phone     string   `json:"phone,omitempty"`
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

// Validate validates the CreateUserRequest
func (r *CreateUserRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return errors.New("username is required")
	}

	if strings.TrimSpace(r.Email) == "" {
		return errors.New("email is required")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}

	if strings.TrimSpace(r.Password) == "" {
		return errors.New("password is required")
	}

	// Validate password strength
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Validate phone format if provided
	if r.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
		if !phoneRegex.MatchString(r.Phone) {
			return errors.New("invalid phone format")
		}
	}

	return nil
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email     *string  `json:"email,omitempty"`
	Phone     *string  `json:"phone,omitempty"`
	FirstName *string  `json:"first_name,omitempty"`
	LastName  *string  `json:"last_name,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

// Validate validates the UpdateUserRequest
func (r *UpdateUserRequest) Validate() error {
	// Validate email format if provided
	if r.Email != nil && *r.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(*r.Email) {
			return errors.New("invalid email format")
		}
	}

	// Validate phone format if provided
	if r.Phone != nil && *r.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
		if !phoneRegex.MatchString(*r.Phone) {
			return errors.New("invalid phone format")
		}
	}

	return nil
}

// ChangePasswordRequest represents the request to change a user's password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Validate validates the ChangePasswordRequest
func (r *ChangePasswordRequest) Validate() error {
	if strings.TrimSpace(r.CurrentPassword) == "" {
		return errors.New("current password is required")
	}

	if strings.TrimSpace(r.NewPassword) == "" {
		return errors.New("new password is required")
	}

	// Validate password strength
	if len(r.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters long")
	}

	return nil
}

// CreateAddressRequest represents the request to create a new address
type CreateAddressRequest struct {
	Address   string `json:"address"`
	Ward      string `json:"ward,omitempty"`
	District  string `json:"district,omitempty"`
	City      string `json:"city"`
	Country   string `json:"country"`
	IsDefault bool   `json:"is_default"`
}

// Validate validates the CreateAddressRequest
func (r *CreateAddressRequest) Validate() error {
	if strings.TrimSpace(r.Address) == "" {
		return errors.New("address is required")
	}

	if strings.TrimSpace(r.City) == "" {
		return errors.New("city is required")
	}

	if strings.TrimSpace(r.Country) == "" {
		return errors.New("country is required")
	}

	return nil
}

// UpdateAddressRequest represents the request to update an address
type UpdateAddressRequest struct {
	Address   *string `json:"address,omitempty"`
	Ward      *string `json:"ward,omitempty"`
	District  *string `json:"district,omitempty"`
	City      *string `json:"city,omitempty"`
	Country   *string `json:"country,omitempty"`
	IsDefault *bool   `json:"is_default,omitempty"`
}

// Validate validates the UpdateAddressRequest
func (r *UpdateAddressRequest) Validate() error {
	// Check if at least one field is provided
	if r.Address == nil && r.Ward == nil && r.District == nil && r.City == nil && r.Country == nil && r.IsDefault == nil {
		return errors.New("at least one field must be provided")
	}

	return nil
}
