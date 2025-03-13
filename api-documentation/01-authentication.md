# Authentication API

The Authentication API provides endpoints for user registration and login.

## Login

Authenticates a user and returns a JWT token for accessing protected endpoints.

### Endpoint

```
POST /api/auth/login
```

### Request Body

```json
{
  "username": "user@example.com",
  "password": "password123"
}
```

| Field    | Type   | Required | Description                                                |
|----------|--------|----------|------------------------------------------------------------|
| username | string | Yes      | User's email address or username                           |
| password | string | Yes      | User's password (minimum 6 characters)                     |

### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Authentication successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "7ad3d8d8-53da-4b48-b1b8-14b09af42630",
    "username": "user",
    "email": "user@example.com",
    "roles": ["admin"]
  }
}
```

### Error Responses

#### Invalid Request (400 Bad Request)

```json
{
  "success": false,
  "message": "Invalid request",
  "error": "Error details"
}
```

#### Validation Failed (400 Bad Request)

```json
{
  "success": false,
  "message": "Validation failed",
  "error": "username is required"
}
```

#### Invalid Credentials (401 Unauthorized)

```json
{
  "success": false,
  "message": "Authentication failed",
  "error": "Invalid credentials"
}
```

#### Server Error (500 Internal Server Error)

```json
{
  "success": false,
  "message": "Authentication failed",
  "error": "Internal server error"
}
```

## Register

Registers a new user in the system.

### Endpoint

```
POST /api/auth/register
```

### Request Body

```json
{
  "email": "newuser@example.com",
  "phone": "1234567890",
  "password": "password123"
}
```

| Field    | Type   | Required | Description                                                |
|----------|--------|----------|------------------------------------------------------------|
| email    | string | Yes*     | User's email address                                       |
| phone    | string | Yes*     | User's phone number                                        |
| password | string | Yes      | User's password (minimum 6 characters)                     |

*Either email or phone is required

### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Registration successful",
  "user_id": "7ad3d8d8-53da-4b48-b1b8-14b09af42630",
  "username": "newuser",
  "email": "newuser@example.com"
}
```

### Error Responses

#### Invalid Request (400 Bad Request)

```json
{
  "success": false,
  "message": "Invalid request",
  "error": "Error details"
}
```

#### Validation Failed (400 Bad Request)

```json
{
  "success": false,
  "message": "Validation failed",
  "error": "email or phone number is required"
}
```

#### User Already Exists (400 Bad Request)

```json
{
  "success": false,
  "message": "Registration failed",
  "error": "Email or phone number already registered"
}
```

#### Server Error (500 Internal Server Error)

```json
{
  "success": false,
  "message": "Registration failed",
  "error": "Internal server error"
}
```

## Implementation Notes for Frontend Developers

### Storing the JWT Token

After a successful login, store the JWT token securely:

```javascript
// Example using localStorage (consider more secure alternatives in production)
const handleLogin = async (username, password) => {
  try {
    const response = await fetch('http://localhost:3000/api/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });
    
    const data = await response.json();
    
    if (data.success) {
      // Store the token
      localStorage.setItem('token', data.token);
      // Store user info
      localStorage.setItem('user', JSON.stringify(data.user));
      return true;
    } else {
      console.error('Login failed:', data.error);
      return false;
    }
  } catch (error) {
    console.error('Login error:', error);
    return false;
  }
};
```

### Using the JWT Token for Authenticated Requests

Include the token in the Authorization header for protected endpoints:

```javascript
const fetchProtectedData = async (url) => {
  const token = localStorage.getItem('token');
  
  if (!token) {
    // Redirect to login or handle unauthenticated state
    return null;
  }
  
  try {
    const response = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });
    
    if (response.status === 401) {
      // Token expired or invalid, redirect to login
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      // Redirect to login page
      return null;
    }
    
    return await response.json();
  } catch (error) {
    console.error('API request error:', error);
    return null;
  }
};
```

### Registration Form Example

```jsx
import React, { useState } from 'react';

const RegistrationForm = () => {
  const [formData, setFormData] = useState({
    email: '',
    phone: '',
    password: '',
  });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    
    try {
      const response = await fetch('http://localhost:3000/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });
      
      const data = await response.json();
      
      if (data.success) {
        setSuccess(true);
        // Redirect to login or automatically log in the user
      } else {
        setError(data.error || 'Registration failed');
      }
    } catch (error) {
      setError('Network error. Please try again.');
    }
  };

  return (
    <div>
      <h2>Register</h2>
      {success ? (
        <div className="success">Registration successful! You can now log in.</div>
      ) : (
        <form onSubmit={handleSubmit}>
          {error && <div className="error">{error}</div>}
          <div>
            <label>Email:</label>
            <input
              type="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
            />
          </div>
          <div>
            <label>Phone (optional):</label>
            <input
              type="text"
              name="phone"
              value={formData.phone}
              onChange={handleChange}
            />
          </div>
          <div>
            <label>Password:</label>
            <input
              type="password"
              name="password"
              value={formData.password}
              onChange={handleChange}
            />
          </div>
          <button type="submit">Register</button>
        </form>
      )}
    </div>
  );
};

export default RegistrationForm;
``` 