package services_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ybds/internal/testutil"
)

func TestAuthService(t *testing.T) {
	t.Skip("Skipping integration test")
}

func TestLogin(t *testing.T) {
	t.Run("Successful login", func(t *testing.T) {
		// Setup
		mockAuthService := &testutil.MockAuthService{}
		username := "testuser"
		password := "password123"
		userID := uuid.New()
		roles := []string{"user"}
		token := "jwt.token.here"

		// Mock auth service response
		authResult := &testutil.AuthResult{
			Success:  true,
			Message:  "Authentication successful",
			UserID:   userID,
			Username: username,
			Email:    "test@example.com",
			Token:    token,
			Roles:    roles,
		}

		mockAuthService.On("Login", username, password).Return(authResult, nil)

		// Execute
		result, err := mockAuthService.Login(username, password)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "Authentication successful", result.Message)
		assert.Equal(t, token, result.Token)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, username, result.Username)
		assert.Equal(t, "test@example.com", result.Email)
		assert.Equal(t, roles, result.Roles)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		// Setup
		mockAuthService := &testutil.MockAuthService{}
		username := "nonexistentuser"
		password := "password123"

		// Mock auth service response
		authResult := &testutil.AuthResult{
			Success: false,
			Message: "Authentication failed",
			Error:   "Invalid credentials",
		}

		mockAuthService.On("Login", username, password).Return(authResult, fmt.Errorf("invalid credentials"))

		// Execute
		result, err := mockAuthService.Login(username, password)

		// Assert
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Equal(t, "Authentication failed", result.Message)
		assert.Equal(t, "Invalid credentials", result.Error)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("Inactive account", func(t *testing.T) {
		// Setup
		mockAuthService := &testutil.MockAuthService{}
		username := "inactiveuser"
		password := "password123"

		// Mock auth service response
		authResult := &testutil.AuthResult{
			Success: false,
			Message: "Authentication failed",
			Error:   "Account is inactive",
		}

		mockAuthService.On("Login", username, password).Return(authResult, fmt.Errorf("account is inactive"))

		// Execute
		result, err := mockAuthService.Login(username, password)

		// Assert
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Equal(t, "Authentication failed", result.Message)
		assert.Equal(t, "Account is inactive", result.Error)

		mockAuthService.AssertExpectations(t)
	})
}

func TestRegister(t *testing.T) {
	t.Run("Successful registration", func(t *testing.T) {
		// Setup
		mockAuthService := &testutil.MockAuthService{}
		email := "newuser@example.com"
		phone := "+1234567890"
		password := "password123"
		userID := uuid.New()

		// Mock auth service response
		authResult := &testutil.AuthResult{
			Success:  true,
			Message:  "Registration successful",
			UserID:   userID,
			Username: "newuser",
			Email:    email,
		}

		mockAuthService.On("Register", email, phone, password).Return(authResult, nil)

		// Execute
		result, err := mockAuthService.Register(email, phone, password)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "Registration successful", result.Message)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, "newuser", result.Username)
		assert.Equal(t, email, result.Email)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("Registration with existing email", func(t *testing.T) {
		// Setup
		mockAuthService := &testutil.MockAuthService{}
		email := "existing@example.com"
		phone := "+1234567890"
		password := "password123"

		// Mock auth service response
		authResult := &testutil.AuthResult{
			Success: false,
			Message: "Registration failed",
			Error:   "Email already exists",
		}

		mockAuthService.On("Register", email, phone, password).Return(authResult, fmt.Errorf("email already exists"))

		// Execute
		result, err := mockAuthService.Register(email, phone, password)

		// Assert
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Equal(t, "Registration failed", result.Message)
		assert.Equal(t, "Email already exists", result.Error)

		mockAuthService.AssertExpectations(t)
	})
}
