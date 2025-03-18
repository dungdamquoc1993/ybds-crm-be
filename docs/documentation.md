# YBDS - Complete Documentation

## Table of Contents
1. [Project Overview](#project-overview)
2. [Features](#features)
3. [Technology Stack](#technology-stack)
4. [Project Structure](#project-structure)
5. [Architecture](#architecture)
6. [Database and Business Rules](#database-and-business-rules)
7. [Getting Started](#getting-started)
8. [Running Tests](#running-tests)
9. [API Documentation](#api-documentation)
10. [Production Deployment](#production-deployment)
11. [Contributing](#contributing)

## Project Overview

This repository contains the backend API for the YBDS application, which serves both a React client and an AI agent.

## Features

- User management (registration, authentication, profile management)
- Product management (inventory, pricing, categories)
- Order processing (creation, status updates, payment processing)
- Notification system (real-time notifications via WebSockets)

## Technology Stack

- Go (Golang)
- Fiber (Web framework)
- GORM (ORM)
- PostgreSQL (Database)
- WebSockets (Real-time communication)
- JWT (Authentication)

## Project Structure

```
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/         # HTTP request handlers
│   │   ├── requests/         # Request models and validation
│   │   └── responses/        # Response models
│   ├── database/             # Database connection and migration
│   ├── models/               # Domain models
│   ├── repositories/         # Data access layer
│   ├── services/             # Business logic
│   └── utils/                # Utility functions
├── pkg/
│   ├── config/               # Configuration management
│   ├── database/             # Database utilities
│   ├── jwt/                  # JWT authentication
│   ├── logger/               # Logging utilities
│   ├── upload/               # File upload utilities
│   └── websocket/            # WebSocket utilities
├── docs/                     # Documentation
├── migrations/               # Database migrations
├── tests/                    # End-to-end tests
│   └── e2e/                  # End-to-end test files
├── scripts/                  # Utility scripts
├── .env                      # Environment variables
├── docker-compose.yml        # Docker Compose configuration
└── Dockerfile                # Docker configuration
```

## Architecture

### Overview

The YBDS application follows a layered architecture that adheres to the Single Responsibility Principle (SRP). Each component has a well-defined responsibility and interacts with other components through clear interfaces.

### Architectural Layers

#### 1. API Layer (Handlers)

**Responsibility**: Handle HTTP requests, validate input, and return appropriate responses.

- **AuthHandler**: Handles authentication and authorization
- **UserHandler**: Handles user management
- **ProductHandler**: Handles product management
- **OrderHandler**: Handles order management
- **NotificationHandler**: Handles notification retrieval and WebSocket connections

#### 2. Service Layer

**Responsibility**: Implement business logic and orchestrate operations across repositories.

- **UserService**: Manages user-related business logic
- **ProductService**: Manages product-related business logic
- **OrderService**: Manages order-related business logic
- **NotificationService**: Manages notification-related business logic
- **JWTService**: Manages JWT token generation and validation

#### 3. Repository Layer

**Responsibility**: Handle data access and persistence.

- **UserRepository**: Manages user data
- **ProductRepository**: Manages product data
- **OrderRepository**: Manages order data
- **NotificationRepository**: Manages notification data

#### 4. Model Layer

**Responsibility**: Define the data structures and business entities.

- **User**: Represents a user in the system
- **Product**: Represents a product
- **Order**: Represents an order
- **Notification**: Represents a notification

### Database Separation

Each domain has its own database to ensure clear separation of concerns:

- **AccountDB**: Stores user-related data
- **ProductDB**: Stores product-related data
- **OrderDB**: Stores order-related data
- **NotificationDB**: Stores notification-related data

### Single Responsibility Principle Implementation

#### Service Independence

Each service is responsible for a specific domain and only interacts with its own repository:

- **UserService** only uses **UserRepository**
- **ProductService** only uses **ProductRepository**
- **OrderService** only uses **OrderRepository**
- **NotificationService** only uses **NotificationRepository**

#### Cross-Service Communication

When a service needs data or functionality from another domain, it calls the appropriate service rather than accessing the repository directly:

- **OrderService** calls **ProductService** to check inventory availability and reserve/release inventory
- **OrderService** calls **UserService** to verify customer existence
- Services call **NotificationService** to create notifications

#### Example: Inventory Management

The inventory management functionality demonstrates SRP:

1. **ProductService** provides methods to manage inventory:
   - `CheckInventoryAvailability`: Checks if there's enough inventory
   - `ReserveInventory`: Reduces inventory quantity
   - `ReleaseInventory`: Increases inventory quantity

2. **OrderService** uses these methods instead of directly accessing the ProductRepository:
   ```go
   // When order is confirmed, reserve inventory
   for _, item := range items {
       if err := s.productService.ReserveInventory(item.InventoryID, item.Quantity); err != nil {
           return err
       }
   }
   ```

### Benefits of This Architecture

1. **Maintainability**: Each component has a clear responsibility, making the code easier to understand and maintain.
2. **Testability**: Components can be tested in isolation by mocking their dependencies.
3. **Scalability**: The clear separation of concerns makes it easier to scale individual components.
4. **Flexibility**: Components can be replaced or modified without affecting the rest of the system.
5. **Microservices Readiness**: The architecture is already organized in a way that would make it easy to extract services into separate microservices in the future.

### Dependency Injection

The application uses constructor-based dependency injection to provide services with their dependencies:

```go
// Initialize services in the correct order to respect dependencies
notificationService := services.NewNotificationService(dbConnections.NotificationDB, hub)
userService := services.NewUserService(dbConnections.AccountDB, notificationService)
productService := services.NewProductService(dbConnections.ProductDB, notificationService)

// Initialize handlers
orderHandler := handlers.NewOrderHandler(dbConnections.OrderDB, productService, userService, notificationService)
```

This approach makes the dependencies explicit and facilitates testing by allowing dependencies to be mocked.

## Database and Business Rules

### Cross-Database Relationships

Since our application uses multiple databases (account, notification, order, product), traditional database foreign key constraints cannot be enforced between databases. The following cross-database relationships must be validated at the application level:

1. **OrderItem to Inventory**: Validate that `InventoryID` in `OrderItem` exists in the `inventory` table in the product database.
2. **Order to User/Guest**: Validate that `CustomerID` in `Order` exists in either the `users` or `guests` table based on the `CustomerType`.
3. **Notification to Recipients**: Validate that `RecipientID` in `Notification` exists in the appropriate table based on the `RecipientType`.

### Business Rules and Validations

#### Order and Payment Processing

##### Payment Method Rules
1. **Cash Payment**:
   - Payment is handled in person
   - No payment tracking system needed
   - Order can proceed to shipping once payment is confirmed
   - Status should be marked as "completed" immediately upon confirmation

2. **Bank Transfer**:
   - Manual verification of bank transfer required
   - Staff should verify the transfer before marking as "completed"
   - Order processing can begin after payment verification

3. **Cash on Delivery (COD)**:
   - No upfront payment required
   - Payment is collected by shipping carrier
   - Carrier handles payment collection and transfer
   - Final payment status updated based on delivery outcome

##### Payment Status Transitions
- **Pending**: Initial state for all orders
- **Completed**:
  - Cash: Set immediately upon staff confirmation
  - Bank Transfer: Set after transfer verification
  - COD: Set after successful delivery and payment collection
- **Failed**:
  - Order is cancelled
  - COD delivery fails
  - Customer refuses delivery/payment
  - Order is returned

##### Order Status Flow
1. **New Order Creation**:
   - Validate customer exists
   - Check inventory availability
   - Calculate total amount
   - Set initial payment status based on payment method

2. **Order Processing**:
   - Confirm payment (except COD)
   - Reserve inventory
   - Create shipment request
   - Update inventory quantities

3. **Order Completion**:
   - Update payment status
   - Release or convert inventory reservation
   - Send notifications

#### Inventory Management

##### Inventory Transaction System
The inventory transaction system tracks all changes to product inventory without requiring client-side input. It is handled automatically by the application layer.

##### Transaction Types
1. **Inbound**:
   - New stock arrivals
   - Returns from customers
   - Inventory adjustments (positive)

2. **Outbound**:
   - Sales
   - Damages
   - Inventory adjustments (negative)

3. **Reservation**:
   - Temporary holds during checkout
   - Released if checkout fails
   - Converted to outbound if order completes

##### Implementation Guidelines
1. **Service Layer Responsibility**:
   ```go
   // Example: Processing an order
   func (s *OrderService) ProcessOrder(order *Order) error {
       tx := s.db.Begin()
       
       // Deduct inventory and log transaction
       for _, item := range order.Items {
           if err := s.inventoryService.DeductStock(item.InventoryID, item.Quantity); err != nil {
               tx.Rollback()
               return err
           }
       }
       
       return tx.Commit()
   }
   ```

2. **Inventory Service**:
   ```go
   func (s *InventoryService) DeductStock(inventoryID uuid.UUID, quantity int) error {
       // Update inventory
       if err := s.updateQuantity(inventoryID, -quantity); err != nil {
           return err
       }
       
       // Log transaction (automatic)
       return s.createTransaction(InventoryTransaction{
           InventoryID: inventoryID,
           Quantity: -quantity,
           Type: TransactionOutbound,
           Reason: ReasonSale,
       })
   }
   ```

#### Other Model Validations

##### Address Model
- Ensure an address is associated with either a User or a Guest, but not both or neither.
- When a User or Guest is deleted, their associated addresses should be handled appropriately.

##### Price Model
- Prevent overlapping price periods for the same product.
- Ensure `EndDate` is after `StartDate` when `EndDate` is provided.

##### Notification System
- Validate recipient existence before sending notifications.
- Implement retry logic for failed notification attempts.

#### Technical Considerations

##### Audit and Security
- Implement comprehensive audit logging for critical operations.
- Use `created_by` and `updated_by` fields to track user actions.
- Implement data versioning for critical entities that require historical tracking.

##### Performance Optimization
- Add application-level caching for frequently accessed data.
- Implement database connection pooling for each database.
- Consider read replicas for high-traffic scenarios.

##### Soft Delete Consistency
- When a parent record is soft-deleted, its related records should not be accessible through the API.
- Repository methods should check if related parent records are not soft-deleted.
- For direct lookups, use JOIN with parent table and check if parent is not deleted.
- For collection lookups, check if parent exists and is not deleted before retrieving the collection.

## Getting Started

### Prerequisites

- Go 1.18 or higher
- PostgreSQL
- Docker (optional, for containerized deployment)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/ybds.git
   cd ybds
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

4. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

## Running Tests

### Unit Tests

To run all unit tests:

```bash
go test ./...
```

To run tests for a specific package:

```bash
go test ./internal/api/handlers
go test ./internal/services
```

To run tests with coverage:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests require a test database. You can set up a test database in your `.env.test` file.

```bash
# Set up test environment
cp .env.example .env.test
# Edit .env.test with test database configuration

# Run integration tests
go test ./tests/integration/...
```

### End-to-End Tests

End-to-end tests simulate client interactions with the API.

```bash
# Run the API in test mode
go run cmd/server/main.go --env=test

# In another terminal, run the E2E tests
go test ./tests/e2e/...
```

### Test Plan

#### Test Priorities

1. **Critical Business Logic**: Focus on testing core business logic in services, especially those that handle financial transactions or inventory management.
2. **API Endpoints**: Ensure all API endpoints work correctly and handle edge cases appropriately.
3. **Database Interactions**: Test repository methods to ensure they interact with the database correctly.
4. **Cross-Service Dependencies**: Verify that services interact with each other correctly through well-defined interfaces.

#### Test Implementation Strategy

1. **Phase 1**: Unit Tests for Core Services
2. **Phase 2**: Repository Tests
3. **Phase 3**: Integration Tests
4. **Phase 4**: End-to-End Tests

#### Test Environment

- **Development**: Local database instances, mock external services, in-memory WebSocket server
- **Staging**: Dedicated test database, sandboxed external services, full WebSocket implementation

#### Continuous Integration

- Run unit tests on every commit
- Run integration tests on pull requests
- Run end-to-end tests before deployment to product
- Generate test coverage reports

## API Documentation

API documentation is available via Swagger UI when the application is running:

```
http://localhost:8080/swagger/index.html
```

## Production Deployment

### Prerequisites

- Docker and Docker Compose installed on the server
- Git installed on the server
- A domain name (optional, but recommended)
- SSL certificate (optional, but recommended)

### Environment Variables

Create a `.env.prod` file in the root directory with the following variables:

```
DB_USER=your_db_user
DB_PASS=your_secure_password
DB_NAME=ybds
JWT_SECRET=your_secure_jwt_secret
```

Make sure to use strong, secure passwords and secrets.

### Deployment Steps

1. Clone the repository:

```bash
git clone https://github.com/yourusername/ybds.git
cd ybds
```

2. Create the necessary directories:

```bash
mkdir -p backups
```

3. Start the application using the production Docker Compose file:

```bash
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

4. Verify that the application is running:

```bash
docker-compose -f docker-compose.prod.yml ps
```

### Image Storage

In the production environment, images are stored in a Docker volume named `uploads-data`. This volume is mounted to the `/app/uploads` directory in the container.

#### Backup Strategy

The production Docker Compose file includes a backup service that creates daily backups of the uploads directory. The backups are stored in the `./backups` directory on the host machine and are kept for 7 days.

To manually trigger a backup:

```bash
docker-compose -f docker-compose.prod.yml exec backup sh -c "tar -czf /backups/uploads-backup-manual-$(date +%Y%m%d-%H%M%S).tar.gz -C /data uploads"
```

#### Restoring from Backup

To restore from a backup:

1. Stop the application:

```bash
docker-compose -f docker-compose.prod.yml down
```

2. Extract the backup to a temporary directory:

```bash
mkdir -p temp
tar -xzf backups/uploads-backup-YYYYMMDD-HHMMSS.tar.gz -C temp
```

3. Start the application:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

4. Copy the extracted files to the container:

```bash
docker cp temp/uploads/. ybds-app:/app/uploads/
```

5. Clean up:

```bash
rm -rf temp
```

### Scaling Considerations

For high-traffic applications, consider the following:

1. **External Storage**: Instead of using a Docker volume, consider using an external storage solution like AWS S3, Google Cloud Storage, or Azure Blob Storage.

2. **CDN**: Use a Content Delivery Network (CDN) to serve images, reducing the load on your application server.

3. **Load Balancing**: Deploy multiple instances of the application behind a load balancer.

### Monitoring

Monitor the disk usage of the uploads directory to ensure you don't run out of space:

```bash
docker exec ybds-app df -h /app/uploads
```

### Troubleshooting

If images are not being served correctly:

1. Check that the uploads directory exists and has the correct permissions:

```bash
docker exec ybds-app ls -la /app/uploads
```

2. Verify that the UPLOAD_DIR environment variable is set correctly:

```bash
docker exec ybds-app env | grep UPLOAD_DIR
```

3. Check the application logs for any errors:

```bash
docker-compose -f docker-compose.prod.yml logs app
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request 