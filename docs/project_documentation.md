# YBDS Project Documentation

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

## Test Plan

### Overview

This test plan outlines a strategy for testing the YBDS application, focusing on ensuring that the application adheres to the Single Responsibility Principle (SRP) and has proper separation of concerns.

### Test Priorities

1. **Critical Business Logic**: Focus on testing core business logic in services, especially those that handle financial transactions or inventory management.
2. **API Endpoints**: Ensure all API endpoints work correctly and handle edge cases appropriately.
3. **Database Interactions**: Test repository methods to ensure they interact with the database correctly.
4. **Cross-Service Dependencies**: Verify that services interact with each other correctly through well-defined interfaces.

### Unit Tests

#### Services

- **UserService**: Test user creation, validation, authentication, role management, password hashing and verification
- **ProductService**: Test product CRUD operations, inventory management, price management
- **OrderService**: Test order creation, validation, status transitions, inventory updates, payment processing
- **NotificationService**: Test notification creation and delivery to different channels

#### Repositories

- **UserRepository**: Test CRUD operations, query filters, pagination
- **ProductRepository**: Test CRUD operations, inventory queries, price queries with date ranges
- **OrderRepository**: Test CRUD operations, order item management, status updates
- **NotificationRepository**: Test notification storage, retrieval, marking as read/unread

#### Handlers

- **AuthHandler**: Test login, registration, token validation, password reset flows
- **UserHandler**: Test user management endpoints, role assignment
- **ProductHandler**: Test product, inventory, and price management endpoints
- **OrderHandler**: Test order creation, management, status updates, payment processing
- **NotificationHandler**: Test notification retrieval, WebSocket connections

### Integration Tests

- **Service-to-Repository Integration**: Test services with their repositories
- **Cross-Service Integration**: Test OrderService with ProductService, OrderService with UserService, Services with NotificationService
- **API Integration**: Test complete API flows for user, product, order, and notification management

### End-to-End Tests

- **User Flows**: Registration, login, profile management, password reset
- **Product Flows**: Product creation, inventory management, price updates
- **Order Flows**: Order creation, checkout, status updates, payment processing, inventory updates
- **Notification Flows**: Real-time notifications, email notifications, preferences

### Test Implementation Strategy

1. **Phase 1**: Unit Tests for Core Services
2. **Phase 2**: Repository Tests
3. **Phase 3**: Integration Tests
4. **Phase 4**: End-to-End Tests

### Test Environment

- **Development**: Local database instances, mock external services, in-memory WebSocket server
- **Staging**: Dedicated test database, sandboxed external services, full WebSocket implementation

### Continuous Integration

- Run unit tests on every commit
- Run integration tests on pull requests
- Run end-to-end tests before deployment to staging
- Generate test coverage reports

### Test Maintenance

- Review and update tests when business logic changes
- Monitor test performance and optimize slow tests
- Regularly review test coverage and identify gaps 