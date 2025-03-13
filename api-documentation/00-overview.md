# YBDS API Documentation

## Overview

This documentation provides comprehensive information about the YBDS API endpoints, request/response formats, and authentication requirements for frontend React.js developers.

## Base URL

All API endpoints are relative to the base URL:

```
http://localhost:3000/api
```

## Authentication

Most endpoints require authentication using JWT (JSON Web Token). After successful login, you'll receive a token that should be included in the `Authorization` header of subsequent requests:

```
Authorization: Bearer <your_token>
```

## Response Format

All API responses follow a consistent format:

### Success Response

```json
{
  "success": true,
  "message": "Operation successful message",
  "data": {
    // Response data specific to the endpoint
  }
}
```

### Error Response

```json
{
  "success": false,
  "message": "Error message",
  "error": "Detailed error description"
}
```

## API Documentation Structure

This documentation is organized by resource type:

1. [Authentication](./01-authentication.md) - Login and registration
2. [Users](./02-users.md) - User management
3. [Products](./03-products.md) - Product management
4. [Inventories](./04-inventories.md) - Inventory management
5. [Prices](./05-prices.md) - Price management
6. [Orders](./06-orders.md) - Order management
7. [Notifications](./07-notifications.md) - Notification management
8. [WebSockets](./08-websockets.md) - Real-time communication

## Status Codes

The API uses standard HTTP status codes:

- `200 OK` - Request succeeded
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required or failed
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Pagination

Endpoints that return collections support pagination with the following query parameters:

- `page` - Page number (default: 1)
- `page_size` - Number of items per page (default: 10)

Paginated responses include metadata:

```json
{
  "success": true,
  "message": "Resources retrieved successfully",
  "data": [...],
  "total": 100,
  "page": 1,
  "page_size": 10,
  "total_pages": 10
}
``` 