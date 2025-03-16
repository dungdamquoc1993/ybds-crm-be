# YBDS API Documentation for React Client Development

## Overview

This document provides a comprehensive guide to the YBDS API for React client development. It includes all API endpoints, request/response formats, and workflow logic to help you easily integrate with the backend.

## Base URL

```
http://localhost:3000/api
```

## Authentication

### Summary
- **Total Endpoints**: 2 (Login, Register)
- **Workflow**: Users register or login to obtain a JWT token, which must be included in all subsequent authenticated requests.
- **Frontend Implementation**: Store the JWT token securely and include it in the Authorization header for all protected API calls.

### Authentication Flow

1. Users register with email/phone and password
2. Users login to get a JWT token
3. The token must be included in the Authorization header for all protected routes
4. Token format: `Bearer {token}`

### Endpoints

#### Register

- **URL**: `/auth/register`
- **Method**: `POST`
- **Auth Required**: No
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "phone": "1234567890",
    "password": "password123"
  }
  ```
  Note: Either email or phone is required, but not both.
- **Response**:
  ```json
  {
    "success": true,
    "message": "Registration successful",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "user@example.com",
    "email": "user@example.com"
  }
  ```

#### Login

- **URL**: `/auth/login`
- **Method**: `POST`
- **Auth Required**: No
- **Request Body**:
  ```json
  {
    "username": "user@example.com",
    "password": "password123"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "Login successful",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "user@example.com",
      "email": "user@example.com",
      "roles": ["user"]
    }
  }
  ```

## Products

### Summary
- **Total Endpoints**: 14 (5 core product endpoints, 4 inventory endpoints, 3 price endpoints, 5 image endpoints)
- **Workflow**: Products are created with basic information, then enhanced with inventories, prices, and images. Each product can have multiple inventories (size/color combinations), prices (with optional end dates), and images (one set as primary).
- **Frontend Implementation**: Create product management interfaces that allow for the complete product lifecycle, from creation to inventory/price management and image uploads.
- **Endpoint Interactions**: Product creation must happen before inventory, price, or image operations. Inventory IDs are used when creating orders.

### Product Workflow

1. Admin/Agent creates products with basic information
2. Products can have multiple inventories (size/color combinations)
3. Products can have multiple prices (with optional end dates)
4. Products can have multiple images (one set as primary)
5. Inventory IDs are referenced when creating orders

### Core Product Endpoints

#### Create Product

- **URL**: `/products`
- **Method**: `POST`
- **Auth Required**: Yes (Admin or Agent)
- **Content-Type**: `multipart/form-data`
- **Request Body**:
  ```
  name: "Product Name"
  description: "Product Description"
  sku: "SKU123"
  category: "Category"
  inventories: [{"size":"M","color":"Red","quantity":10,"location":"Warehouse A"}]
  prices: [{"price":99.99,"currency":"USD","end_date":"2023-12-31T23:59:59Z"}]
  images: (file uploads)
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "Product created successfully",
    "data": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Product Name",
      "description": "Product Description",
      "sku": "SKU123",
      "category": "Category",
      "image_url": "/uploads/products/product_image.jpg",
      "inventories": [...],
      "prices": [...],
      "images": [...],
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-01T12:00:00Z"
    }
  }
  ```

#### Get Products

- **URL**: `/products`
- **Method**: `GET`
- **Auth Required**: Yes (Admin or Agent)
- **Query Parameters**:
  - `page`: Page number (default: 1)
  - `page_size`: Items per page (default: 10)
  - `search`: Search term
  - `category`: Filter by category
- **Response**:
  ```json
  {
    "success": true,
    "message": "Products retrieved successfully",
    "products": [...],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_pages": 10
  }
  ```

#### Get Product by ID

- **URL**: `/products/:id`
- **Method**: `GET`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Detailed product information including inventories, prices, and images

#### Update Product

- **URL**: `/products/:id`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**: Product information to update
- **Response**: Updated product information

#### Delete Product

- **URL**: `/products/:id`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Success message

### Product Inventory Endpoints

#### Create Inventory

- **URL**: `/products/:id/inventories`
- **Method**: `POST`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "size": "M",
    "color": "Red",
    "quantity": 10,
    "location": "Warehouse A"
  }
  ```
- **Response**: Created inventory information

#### Create Multiple Inventories

- **URL**: `/products/:id/inventories/batch`
- **Method**: `POST`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "inventories": [
      {
        "size": "M",
        "color": "Red",
        "quantity": 10,
        "location": "Warehouse A"
      },
      {
        "size": "L",
        "color": "Blue",
        "quantity": 5,
        "location": "Warehouse B"
      }
    ]
  }
  ```
- **Response**: Success message

#### Update Inventory

- **URL**: `/products/inventories/:id`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "size": "M",
    "color": "Red",
    "quantity": 15,
    "location": "Warehouse A"
  }
  ```
- **Response**: Updated inventory information

#### Delete Inventory

- **URL**: `/products/inventories/:id`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Success message

### Product Price Endpoints

#### Create Price

- **URL**: `/products/:id/prices`
- **Method**: `POST`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "price": 99.99,
    "currency": "USD",
    "end_date": "2023-12-31T23:59:59Z"
  }
  ```
- **Response**: Created price information

#### Update Price

- **URL**: `/products/prices/:id`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "price": 89.99,
    "currency": "USD",
    "end_date": "2023-12-31T23:59:59Z"
  }
  ```
- **Response**: Updated price information

#### Delete Price

- **URL**: `/products/prices/:id`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Success message

### Product Image Endpoints

#### Get Product Images

- **URL**: `/products/:id/images`
- **Method**: `GET`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: List of product images

#### Upload Product Image

- **URL**: `/products/:id/images`
- **Method**: `POST`
- **Auth Required**: Yes (Admin or Agent)
- **Content-Type**: `multipart/form-data`
- **Request Body**: Image file
- **Response**: Uploaded image information

#### Set Primary Product Image

- **URL**: `/products/:id/images/:imageId/primary`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Updated image information

#### Reorder Product Images

- **URL**: `/products/:id/images/reorder`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "imageIds": ["id1", "id2", "id3"]
  }
  ```
- **Response**: Success message

#### Delete Product Image

- **URL**: `/products/:id/images/:imageId`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Success message

## Orders

### Summary
- **Total Endpoints**: 10 (7 core order endpoints, 3 order item endpoints)
- **Workflow**: Orders are created with customer information and items, then progress through various status stages. Order items are linked to product inventories, and inventory quantities are automatically adjusted.
- **Frontend Implementation**: Create order management interfaces that handle the complete order lifecycle, from creation to fulfillment and shipping.
- **Endpoint Interactions**: Orders reference inventory IDs from the product system. Order status changes trigger notifications via WebSocket. Shipment information can be added to orders.

### Order Workflow

1. Admin/Agent creates an order with customer information and items (referencing inventory IDs)
2. Order status progresses through various stages (pending_confirmation, confirmed, shipped, delivered, etc.)
3. Order items are linked to product inventories
4. Inventory quantities are automatically adjusted when orders are created or modified
5. Notifications are sent to relevant users when order status changes
6. Shipment information can be added to track the order

### Core Order Endpoints

#### Create Order

- **URL**: `/orders`
- **Method**: `POST`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "payment_method": "cash",
    "status": "pending_confirmation",
    "notes": "Please deliver in the morning",
    "discount_amount": 10.50,
    "discount_reason": "Loyalty discount",
    "items": [
      {
        "inventory_id": "550e8400-e29b-41d4-a716-446655440000",
        "quantity": 2
      }
    ],
    "shipping_address": "123 Main St",
    "shipping_ward": "Ward 1",
    "shipping_district": "District 1",
    "shipping_city": "Ho Chi Minh City",
    "shipping_country": "Vietnam",
    "customer_name": "John Doe",
    "customer_email": "john@example.com",
    "customer_phone": "1234567890"
  }
  ```
- **Response**: Created order information

#### Get Orders

- **URL**: `/orders`
- **Method**: `GET`
- **Auth Required**: Yes (Admin or Agent)
- **Query Parameters**:
  - `page`: Page number (default: 1)
  - `page_size`: Items per page (default: 10)
  - `status`: Filter by status
  - `search`: Search term
  - `start_date`: Filter by start date
  - `end_date`: Filter by end date
- **Response**: List of orders with pagination

#### Get Order by ID

- **URL**: `/orders/:id`
- **Method**: `GET`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Detailed order information including items and shipment

#### Update Order Status

- **URL**: `/orders/:id/status`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin only)
- **Request Body**:
  ```json
  {
    "status": "confirmed"
  }
  ```
- **Response**: Updated order information
- **Side Effects**: Triggers notifications via WebSocket

#### Update Order Details

- **URL**: `/orders/:id/details`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "payment_method": "credit_card",
    "notes": "Updated notes",
    "discount_amount": 15.00,
    "discount_reason": "Special discount",
    "shipping_address": "456 New St",
    "shipping_ward": "Ward 2",
    "shipping_district": "District 2",
    "shipping_city": "Ho Chi Minh City",
    "shipping_country": "Vietnam",
    "customer_name": "John Doe",
    "customer_email": "john@example.com",
    "customer_phone": "1234567890"
  }
  ```
- **Response**: Updated order information

#### Update Shipment

- **URL**: `/orders/:id/shipment`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "tracking_number": "TRACK123456",
    "carrier": "DHL"
  }
  ```
- **Response**: Updated order information with shipment details
- **Side Effects**: May trigger status change and notifications

#### Delete Order

- **URL**: `/orders/:id`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin only)
- **Response**: Success message
- **Side Effects**: Inventory quantities may be adjusted

### Order Item Endpoints

#### Add Order Item

- **URL**: `/orders/:id/items`
- **Method**: `POST`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "inventory_id": "550e8400-e29b-41d4-a716-446655440000",
    "quantity": 2
  }
  ```
- **Response**: Added item information
- **Side Effects**: Inventory quantities are adjusted

#### Update Order Item

- **URL**: `/orders/items/:id`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Request Body**:
  ```json
  {
    "quantity": 3
  }
  ```
- **Response**: Updated item information
- **Side Effects**: Inventory quantities are adjusted

#### Delete Order Item

- **URL**: `/orders/items/:id`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Success message
- **Side Effects**: Inventory quantities are adjusted

## Notifications

### Summary
- **Total Endpoints**: 4
- **Workflow**: Notifications are generated automatically for various events and can be viewed, marked as read individually, or all at once. Real-time notifications are delivered via WebSocket.
- **Frontend Implementation**: Create a notification system that displays notifications to users and updates in real-time via WebSocket.
- **Endpoint Interactions**: Notifications are triggered by actions in other systems (order status changes, etc.) and delivered via both API and WebSocket.

### Notification Workflow

1. Notifications are generated automatically for various events (order creation, status changes, etc.)
2. Users can view all notifications or only unread ones
3. Users can mark notifications as read individually or all at once
4. Real-time notifications are delivered via WebSocket

### Endpoints

#### Get Notifications

- **URL**: `/notifications`
- **Method**: `GET`
- **Auth Required**: Yes (Admin or Agent)
- **Query Parameters**:
  - `page`: Page number (default: 1)
  - `page_size`: Items per page (default: 10)
  - `unread_only`: Filter by unread status (default: false)
- **Response**: List of notifications with pagination

#### Get Unread Notifications

- **URL**: `/notifications/unread`
- **Method**: `GET`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: List of unread notifications

#### Mark Notification as Read

- **URL**: `/notifications/:id/read`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Updated notification information

#### Mark All Notifications as Read

- **URL**: `/notifications/read-all`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin or Agent)
- **Response**: Success message

## WebSocket

### Summary
- **Total Endpoints**: 1
- **Workflow**: WebSocket provides real-time communication for notifications and order status changes.
- **Frontend Implementation**: Establish a WebSocket connection to receive real-time updates without polling the server.
- **Endpoint Interactions**: WebSocket events are triggered by actions in other systems (order status changes, etc.).

### WebSocket Connection

- **URL**: `/ws`
- **Auth Required**: Yes (JWT token in query parameter or header)
- **Connection Process**:
  1. Connect to the WebSocket endpoint with the JWT token
  2. Listen for events from the server
  3. Handle events appropriately in the UI

### WebSocket Events

#### Notification Event
```json
{
  "type": "notification",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "title": "New Order",
    "message": "A new order has been created",
    "read": false,
    "created_at": "2023-01-01T12:00:00Z"
  }
}
```

#### Order Status Change Event
```json
{
  "type": "order_status_change",
  "data": {
    "order_id": "550e8400-e29b-41d4-a716-446655440000",
    "old_status": "pending_confirmation",
    "new_status": "confirmed",
    "timestamp": "2023-01-01T12:00:00Z"
  }
}
```

## File Uploads

### Summary
- **Total Endpoints**: 1 (Product Image Upload)
- **Workflow**: Files are uploaded using multipart/form-data and stored on the server.
- **Frontend Implementation**: Use form data to upload files and handle the response appropriately.
- **Endpoint Interactions**: Uploaded files are referenced by other systems (products, etc.).

### Image Upload Process

1. Images are uploaded using multipart/form-data
2. The server stores the images in the `/uploads/products/` directory
3. The server returns the image URL and other metadata
4. Images can be accessed directly via their URL

### Upload Endpoint

#### Upload Product Image

- **URL**: `/products/:id/images`
- **Method**: `POST`
- **Auth Required**: Yes (Admin or Agent)
- **Content-Type**: `multipart/form-data`
- **Request Body**:
  ```
  image: (file)
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "Image uploaded successfully",
    "data": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "product_id": "550e8400-e29b-41d4-a716-446655440001",
      "url": "/uploads/products/product_image.jpg",
      "filename": "product_image.jpg",
      "is_primary": false,
      "sort_order": 0,
      "created_at": "2023-01-01T12:00:00Z",
      "updated_at": "2023-01-01T12:00:00Z"
    }
  }
  ```

## Frontend Implementation Guide

### Authentication Flow
1. Create login/register forms with validation
2. Store JWT token securely (localStorage or secure cookie)
3. Create an axios interceptor to add the token to all requests
4. Implement logout functionality to clear the token

### Product Management
1. Create a product listing page with filters and pagination
2. Implement a product detail page showing all information
3. Create forms for adding/editing products, inventories, prices, and images
4. Use a file upload component for product images
5. Implement inventory management with size/color combinations

### Order Management
1. Create an order listing page with filters and pagination
2. Implement an order detail page showing all information
3. Create forms for adding/editing orders and order items
4. Implement order status workflow with appropriate UI indicators
5. Add shipment tracking information display

### Notification System
1. Create a notification component that shows unread notifications
2. Implement WebSocket connection for real-time notifications
3. Add ability to mark notifications as read
4. Display notification badges and counters

### WebSocket Integration
1. Establish WebSocket connection on application startup
2. Handle reconnection if the connection is lost
3. Process incoming events and update the UI accordingly
4. Implement event handlers for different event types

### File Upload Implementation
1. Create a file upload component with preview
2. Implement drag-and-drop functionality
3. Show upload progress indicators
4. Handle upload errors gracefully

## Recommended React Libraries

1. **API Requests**: axios, react-query
2. **State Management**: Redux Toolkit, Zustand
3. **Form Handling**: Formik, React Hook Form
4. **Validation**: Yup, Zod
5. **UI Components**: Material-UI, Ant Design, Chakra UI
6. **WebSocket**: socket.io-client, use-websocket
7. **File Upload**: react-dropzone, uppy
8. **Routing**: react-router-dom

## Conclusion

This API documentation provides a comprehensive guide to integrating your React client with the YBDS backend. By following the workflows and endpoint specifications, you should be able to create a fully functional client application that interacts seamlessly with the backend services.

The key to a successful implementation is understanding how the different endpoints interact with each other and how to handle the data flow in your React application. By leveraging the WebSocket connection for real-time updates and properly managing the authentication flow, you can create a responsive and user-friendly application.
