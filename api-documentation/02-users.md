# Users API

The Users API provides endpoints for retrieving user information. All endpoints in this section require authentication.

## Get All Users

Retrieves a paginated list of all users.

### Endpoint

```
GET /api/users
```

### Query Parameters

| Parameter | Type   | Required | Description                                |
|-----------|--------|----------|--------------------------------------------|
| page      | number | No       | Page number (default: 1)                   |
| page_size | number | No       | Number of items per page (default: 10)     |

### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": {
    "users": [
      {
        "id": "7ad3d8d8-53da-4b48-b1b8-14b09af42630",
        "username": "admin",
        "email": "admin@example.com",
        "first_name": "Admin",
        "last_name": "User",
        "phone": "1234567890",
        "status": "active",
        "created_at": "2023-01-01T00:00:00Z",
        "updated_at": "2023-01-01T00:00:00Z"
      },
      // More users...
    ],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_pages": 10
  }
}
```

### Error Response (500 Internal Server Error)

```json
{
  "success": false,
  "message": "Failed to retrieve users",
  "error": "Error details"
}
```

## Get User by ID

Retrieves detailed information about a specific user.

### Endpoint

```
GET /api/users/:id
```

### Path Parameters

| Parameter | Type   | Required | Description                                |
|-----------|--------|----------|--------------------------------------------|
| id        | string | Yes      | User ID (UUID format)                      |

### Success Response (200 OK)

```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": "7ad3d8d8-53da-4b48-b1b8-14b09af42630",
    "username": "admin",
    "email": "admin@example.com",
    "first_name": "Admin",
    "last_name": "User",
    "phone": "1234567890",
    "status": "active",
    "roles": ["admin"],
    "addresses": [
      {
        "id": "8bd3d8d8-53da-4b48-b1b8-14b09af42631",
        "street": "123 Main St",
        "city": "Anytown",
        "state": "CA",
        "postal_code": "12345",
        "country": "USA",
        "is_default": true
      }
    ],
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Error Responses

#### Invalid User ID Format (400 Bad Request)

```json
{
  "success": false,
  "message": "Invalid user ID format",
  "error": "Error details"
}
```

#### User Not Found (404 Not Found)

```json
{
  "success": false,
  "message": "User not found",
  "error": "Error details"
}
```

## Get Guest by ID

Retrieves detailed information about a guest user.

### Endpoint

```
GET /api/guests/:id
```

### Path Parameters

| Parameter | Type   | Required | Description                                |
|-----------|--------|----------|--------------------------------------------|
| id        | string | Yes      | Guest ID (UUID format)                     |

### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Guest retrieved successfully",
  "data": {
    "id": "9cd3d8d8-53da-4b48-b1b8-14b09af42632",
    "email": "guest@example.com",
    "phone": "1234567890",
    "first_name": "Guest",
    "last_name": "User",
    "addresses": [
      {
        "id": "8bd3d8d8-53da-4b48-b1b8-14b09af42631",
        "street": "123 Main St",
        "city": "Anytown",
        "state": "CA",
        "postal_code": "12345",
        "country": "USA",
        "is_default": true
      }
    ],
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Error Responses

#### Invalid Guest ID Format (400 Bad Request)

```json
{
  "success": false,
  "message": "Invalid guest ID format",
  "error": "Error details"
}
```

#### Guest Not Found (404 Not Found)

```json
{
  "success": false,
  "message": "Guest not found",
  "error": "Error details"
}
```

## Implementation Notes for Frontend Developers

### User Profile Component Example

```jsx
import React, { useState, useEffect } from 'react';

const UserProfile = ({ userId }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchUserData = async () => {
      setLoading(true);
      setError('');
      
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          setError('Authentication required');
          setLoading(false);
          return;
        }
        
        const response = await fetch(`http://localhost:3000/api/users/${userId}`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });
        
        if (response.status === 401) {
          setError('Session expired. Please log in again.');
          localStorage.removeItem('token');
          // Redirect to login
          setLoading(false);
          return;
        }
        
        const data = await response.json();
        
        if (data.success) {
          setUser(data.data);
        } else {
          setError(data.error || 'Failed to load user data');
        }
      } catch (error) {
        setError('Network error. Please try again.');
      } finally {
        setLoading(false);
      }
    };
    
    if (userId) {
      fetchUserData();
    }
  }, [userId]);

  if (loading) return <div>Loading...</div>;
  if (error) return <div className="error">{error}</div>;
  if (!user) return <div>No user data available</div>;

  return (
    <div className="user-profile">
      <h2>{user.first_name} {user.last_name}</h2>
      <div className="user-details">
        <p><strong>Email:</strong> {user.email}</p>
        <p><strong>Username:</strong> {user.username}</p>
        <p><strong>Phone:</strong> {user.phone}</p>
        <p><strong>Status:</strong> {user.status}</p>
        <p><strong>Roles:</strong> {user.roles.join(', ')}</p>
      </div>
      
      {user.addresses && user.addresses.length > 0 && (
        <div className="user-addresses">
          <h3>Addresses</h3>
          {user.addresses.map(address => (
            <div key={address.id} className="address-card">
              <p>{address.street}</p>
              <p>{address.city}, {address.state} {address.postal_code}</p>
              <p>{address.country}</p>
              {address.is_default && <span className="default-badge">Default</span>}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default UserProfile;
```

### Users List Component Example

```jsx
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

const UsersList = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      setError('');
      
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          setError('Authentication required');
          setLoading(false);
          return;
        }
        
        const response = await fetch(
          `http://localhost:3000/api/users?page=${page}&page_size=${pageSize}`, 
          {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }
        );
        
        if (response.status === 401) {
          setError('Session expired. Please log in again.');
          localStorage.removeItem('token');
          // Redirect to login
          setLoading(false);
          return;
        }
        
        const data = await response.json();
        
        if (data.success) {
          setUsers(data.data.users);
          setTotalPages(data.data.total_pages);
        } else {
          setError(data.error || 'Failed to load users');
        }
      } catch (error) {
        setError('Network error. Please try again.');
      } finally {
        setLoading(false);
      }
    };
    
    fetchUsers();
  }, [page, pageSize]);

  const handlePageChange = (newPage) => {
    if (newPage >= 1 && newPage <= totalPages) {
      setPage(newPage);
    }
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div className="error">{error}</div>;

  return (
    <div className="users-list">
      <h2>Users</h2>
      
      <table className="users-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {users.map(user => (
            <tr key={user.id}>
              <td>{user.first_name} {user.last_name}</td>
              <td>{user.email}</td>
              <td>{user.status}</td>
              <td>
                <Link to={`/users/${user.id}`}>View Details</Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      
      <div className="pagination">
        <button 
          onClick={() => handlePageChange(page - 1)} 
          disabled={page === 1}
        >
          Previous
        </button>
        <span>Page {page} of {totalPages}</span>
        <button 
          onClick={() => handlePageChange(page + 1)} 
          disabled={page === totalPages}
        >
          Next
        </button>
      </div>
    </div>
  );
};

export default UsersList; 