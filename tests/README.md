# YBDS API Tests

This directory contains tests for the YBDS API. The tests are organized into different categories:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test the interaction between components
- **End-to-End Tests**: Test the entire application flow

## Running Tests

### Unit Tests

To run all unit tests:

```bash
go test ./internal/... -v
```

To run tests for a specific package:

```bash
go test ./internal/api/handlers -v
```

### End-to-End Tests

End-to-end tests require a running instance of the API. These tests simulate real user interactions with the API.

To run the end-to-end tests:

1. Start the API server:

```bash
go run cmd/app/main.go
```

2. In a separate terminal, run the end-to-end tests:

```bash
go test ./tests/e2e -v
```

You can configure the end-to-end tests using environment variables:

- `TEST_API_URL`: The base URL of the API (default: `http://localhost:8080`)
- `TEST_USERNAME`: The username to use for authentication (default: `admin@example.com`)
- `TEST_PASSWORD`: The password to use for authentication (default: `password123`)

Example:

```bash
TEST_API_URL=http://localhost:8080 TEST_USERNAME=admin@example.com TEST_PASSWORD=password123 go test ./tests/e2e -v
```

## Test Coverage

To generate a test coverage report:

```bash
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

This will open a browser window showing the test coverage for each file.

## Writing Tests

### Unit Tests

Unit tests should be placed in the same package as the code they're testing, with a `_test.go` suffix. For example, tests for `internal/api/handlers/product_handler.go` should be in `internal/api/handlers/product_handler_test.go`.

Use the `testutil` package for common testing utilities:

```go
import (
    "testing"
    "github.com/ybds/internal/testutil"
)

func TestSomething(t *testing.T) {
    // Use test utilities
    app := testutil.SetupTestApp()
    // ...
}
```

### End-to-End Tests

End-to-end tests should be placed in the `tests/e2e` directory. These tests should use the `TestClient` to interact with the API.

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
    client := NewTestClient("http://localhost:8080")
    // Authenticate
    err := client.Login("admin@example.com", "password123")
    assert.NoError(t, err)
    
    // Make API requests
    resp, err := client.SendRequest("GET", "/api/products", nil, nil)
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    // ...
}
``` 