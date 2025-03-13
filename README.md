# YBDS API

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
│   └── app/
│       └── main.go           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/         # HTTP request handlers
│   │   ├── requests/         # Request models and validation
│   │   └── responses/        # Response models
│   ├── config/               # Application configuration
│   ├── database/             # Database connection and migrations
│   ├── middleware/           # HTTP middleware
│   ├── models/               # Data models
│   ├── repositories/         # Data access layer
│   ├── services/             # Business logic
│   └── testutil/             # Testing utilities
├── pkg/
│   ├── upload/               # File upload utilities
│   └── websocket/            # WebSocket implementation
└── tests/                    # End-to-end tests
```

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
   go run cmd/app/main.go
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
go run cmd/app/main.go --env=test

# In another terminal, run the E2E tests
go test ./tests/e2e/...
```

## API Documentation

API documentation is available via Swagger UI when the application is running:

```
http://localhost:8080/swagger/index.html
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Testing

### Testing Approach

The application includes comprehensive tests for all handlers and API endpoints:

1. **Unit Tests**: Tests for individual handlers to verify their functionality in isolation.
2. **Integration Tests**: Tests for API endpoints to verify the interaction between components.
3. **End-to-End Tests**: Tests that simulate client interactions with the API.

### Test Structure

The tests are organized as follows:

- **Handler Tests**: Located in `internal/api/handlers/*_test.go` files, these tests verify the functionality of individual handlers.
- **Test Utilities**: Located in `internal/testutil/` directory, these utilities provide common functionality for testing.

### Mock Services

Each handler test uses its own mock implementation of the services it depends on:

- **Mock Order Service**: Used in order handler tests to simulate order-related operations.
- **Mock User Service**: Used in user handler tests to simulate user-related operations.
- **Mock Product Service**: Used in product handler tests to simulate product-related operations.
- **Mock Notification Service**: Each handler has its own mock notification service implementation.

### Running Tests

To run all tests:

```bash
go test ./...
```

To run tests for a specific package:

```bash
go test ./internal/api/handlers
```

To run a specific test:

```bash
go test ./internal/api/handlers -run TestOrderHandler
``` 