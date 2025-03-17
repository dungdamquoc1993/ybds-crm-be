# YBDS API

This repository contains the backend API for the YBDS application, which serves both a React client and an AI agent.

## Quick Start

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

## Docker Setup

Run the application using Docker Compose:

```bash
# Development
docker-compose up -d

# Production
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

## API Documentation

API documentation is available via Swagger UI when the application is running:

```
http://localhost:8080/swagger/index.html
```

## Comprehensive Documentation

For detailed documentation, please refer to the [Complete Documentation](docs/documentation.md) which includes:

- Project Architecture
- Database and Business Rules
- Testing Strategy
- Production Deployment Guide
- And more...

## License

This project is licensed under the MIT License - see the LICENSE file for details.

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
│   ├── swagger.json          # Swagger API documentation
│   ├── swagger.yaml          # Swagger API documentation
│   ├── docs.go               # Swagger generated code
│   ├── project_documentation.md # Project architecture and test plan
│   └── database_and_business_rule_for_application.md # Database and business rules
├── migrations/               # Database migrations
├── tests/                    # End-to-end tests
│   └── e2e/                  # End-to-end test files
├── scripts/                  # Utility scripts
├── .env                      # Environment variables
├── docker-compose.yml        # Docker Compose configuration
└── Dockerfile                # Docker configuration
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

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

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

# Real-time Notifications System

## Overview

This document provides comprehensive guidelines for the real-time notifications system, including backend WebSocket implementation and frontend integration. The WebSocket service enables real-time delivery of notifications related to orders, products, inventory, and user activities.

## WebSocket Implementation

### Backend Configuration

The backend WebSocket implementation uses the Fiber WebSocket package with JWT authentication:

```go
// WebSocket handler with JWT auth
wsHandler := pkgws.NewHandler(hub, pkgws.JWTAuthFunc(
    // Function to extract token from request
    func(c *fiber.Ctx) string {
        token := c.Query("token")
        if token != "" {
            fmt.Printf("[WebSocket] Token from query: %s\n", token[:10]+"...")
        } else {
            fmt.Printf("[WebSocket] No token provided in query\n")
        }
        return token
    },
    // Function to validate token
    func(tokenString string) (string, []string, error) {
        // Validate the JWT token
        claims, err := jwtService.ValidateToken(tokenString)
        if err != nil {
            fmt.Printf("[WebSocket] Token validation error: %v\n", err)
            return "", nil, err
        }
        
        fmt.Printf("[WebSocket] Token validated successfully for user %s with roles %v\n", claims.UserID, claims.Roles)
        return claims.UserID, claims.Roles, nil
    },
))

// Register WebSocket routes
wsGroup := api.Group("/ws")
wsGroup.Use(wsHandler.Middleware())
wsGroup.Get("/", fiberwsocket.New(wsHandler.HandleConnection))
```

### Frontend Integration

#### 1. Install Required Packages

```bash
npm install react-use-websocket
```

#### 2. Create a WebSocket Context

Create a file `src/contexts/WebSocketContext.tsx`:

```tsx
import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import useWebSocket from 'react-use-websocket';
import { useDispatch } from 'react-redux';
import { addNotification } from '../store/features/notification/slice';
import type { Notification as NotificationType } from '../store/features/notification/types';
import { useAuth } from '../hooks/useAuth';

interface WebSocketContextType {
  isConnected: boolean;
  lastMessage: MessageEvent | null;
  sendMessage: (message: string) => void;
  messageHistory: WebSocketMessage[];
  clearHistory: () => void;
  connectionAttempts: number;
}

interface WebSocketMessage {
  type: string;
  payload: WebSocketNotificationPayload;
  timestamp: number;
}

interface WebSocketNotificationPayload {
  id: string;
  user_id?: string;
  event_type?: string;
  title?: string;
  message?: string;
  created_at?: string;
  updated_at?: string;
  priority?: 'low' | 'normal' | 'high' | 'urgent';
  metadata?: {
    id?: string;
    product_id?: string;
    order_id?: string;
    [key: string]: unknown;
  };
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

// Maximum number of messages to keep in history
const MAX_HISTORY_LENGTH = 50;

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isConnected, setIsConnected] = useState(false);
  const [messageHistory, setMessageHistory] = useState<WebSocketMessage[]>([]);
  const [connectionAttempts, setConnectionAttempts] = useState(0);
  const dispatch = useDispatch();
  const { token } = useAuth();
  
  // Determine WebSocket URL from environment or default
  // IMPORTANT: Include the /api prefix for the WebSocket URL
  const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:3000/api/ws';
  
  // Create connection URL with token
  const connectionUrl = token ? `${wsUrl}?token=${token}` : null;
  
  // Debug log the connection URL (without exposing full token)
  useEffect(() => {
    if (connectionUrl) {
      const tokenPreview = token ? `${token.substring(0, 10)}...` : 'none';
      console.log(`WebSocket attempting connection to: ${wsUrl}`);
      console.log(`Using token: ${tokenPreview}`);
    } else {
      console.log('WebSocket not connecting - no token available');
    }
  }, [connectionUrl, token, wsUrl]);
  
  // Setup WebSocket connection
  const { lastMessage, sendMessage, readyState } = useWebSocket(connectionUrl, {
    onOpen: () => {
      console.log('✅ WebSocket connection established successfully');
      setIsConnected(true);
      setConnectionAttempts(0);
    },
    onClose: (event) => {
      console.log(`🔴 WebSocket connection closed with code ${event.code}, reason: ${event.reason || 'No reason provided'}`);
      setIsConnected(false);
    },
    onError: (event) => {
      console.error('❌ WebSocket error:', event);
      setConnectionAttempts(prev => prev + 1);
    },
    onReconnectStop: (numAttempts) => {
      console.error(`❌ WebSocket reconnection stopped after ${numAttempts} attempts`);
    },
    shouldReconnect: (closeEvent) => {
      // Don't keep trying if we get a clear auth error
      if (closeEvent.code === 1008) {
        console.error('🔑 WebSocket authentication failed - check your token');
        return false;
      }
      return true;
    },
    reconnectAttempts: 10,
    reconnectInterval: (attemptNumber) => Math.min(Math.pow(2, attemptNumber) * 1000, 30000), // Exponential backoff
    retryOnError: true,
  });
  
  // Clear message history
  const clearHistory = useCallback(() => {
    setMessageHistory([]);
  }, []);
  
  // Process incoming WebSocket messages
  useEffect(() => {
    if (lastMessage) {
      try {
        console.log('📨 Received WebSocket message:', lastMessage.data);
        const data = JSON.parse(lastMessage.data);
        
        // Store message in history
        const message: WebSocketMessage = {
          type: data.type,
          payload: data.payload,
          timestamp: Date.now(),
        };
        
        setMessageHistory((prev) => {
          const updatedHistory = [message, ...prev];
          // Limit history length
          return updatedHistory.slice(0, MAX_HISTORY_LENGTH);
        });
        
        // Process notifications
        if (data.type === 'notification') {
          console.log('📣 Processing notification:', data.payload);
          // Process the notification based on its type
          const processedNotification = processNotification(data.payload);
          
          // Add the notification to the Redux store
          dispatch(addNotification(processedNotification));
          
          // Show desktop notification for high priority messages
          if (data.payload.priority === 'high' || data.payload.priority === 'urgent') {
            showDesktopNotification(data.payload);
          }
        }
      } catch (e) {
        console.error('❌ Failed to parse WebSocket message:', e);
      }
    }
  }, [lastMessage, dispatch]);

  // Process different notification types
  const processNotification = (payload: WebSocketNotificationPayload): NotificationType => {
    // Ensure we have a properly formatted notification
    const notification: NotificationType = {
      id: payload.id,
      user_id: payload.user_id || '',
      type: payload.event_type || 'system',
      title: payload.title || 'Notification',
      message: payload.message || '',
      is_read: false,
      redirect_url: getRedirectUrl(payload),
      created_at: payload.created_at || new Date().toISOString(),
      updated_at: payload.updated_at || new Date().toISOString(),
    };

    return notification;
  };

  // Show a desktop notification
  const showDesktopNotification = (payload: WebSocketNotificationPayload) => {
    if (!('Notification' in window)) {
      console.log('This browser does not support desktop notifications');
      return;
    }
    
    if (Notification.permission === 'granted') {
      new Notification(payload.title || 'New Notification', {
        body: payload.message,
        icon: '/logo192.png', // Make sure this path is correct
      });
    } else if (Notification.permission !== 'denied') {
      Notification.requestPermission().then(permission => {
        if (permission === 'granted') {
          new Notification(payload.title || 'New Notification', {
            body: payload.message,
            icon: '/logo192.png',
          });
        }
      });
    }
  };

  // Get appropriate redirect URL based on notification type
  const getRedirectUrl = (payload: WebSocketNotificationPayload): string => {
    switch (payload.event_type) {
      case 'order_created':
      case 'order_status_changed':
      case 'order_payment_received':
      case 'order_shipped':
      case 'order_delivered':
      case 'order_cancelled':
        return `/orders/${payload.metadata?.id || ''}`;
      
      case 'product_created':
      case 'product_updated':
      case 'product_deleted':
      case 'product_price_changed':
        return `/products/${payload.metadata?.id || ''}`;
      
      case 'low_stock':
      case 'out_of_stock':
      case 'back_in_stock':
      case 'inventory_updated':
        return `/products/${payload.metadata?.product_id || ''}`;
      
      case 'user_registered':
      case 'user_login':
      case 'password_changed':
      case 'user_role_changed':
        return `/users/${payload.metadata?.id || ''}`;
      
      default:
        return '';
    }
  };
  
  // Re-establish connection when token changes
  useEffect(() => {
    if (token) {
      console.log('🔄 Token changed, reconnecting WebSocket');
      // The connection will be re-established automatically when the URL changes
    }
  }, [token]);
  
  return (
    <WebSocketContext.Provider value={{ 
      isConnected, 
      lastMessage, 
      sendMessage,
      messageHistory,
      clearHistory,
      connectionAttempts
    }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocketContext = (): WebSocketContextType => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider');
  }
  return context;
};
```

#### 3. Wrap Your App with the WebSocket Provider

In your `src/App.tsx`:

```tsx
import React from 'react';
import { WebSocketProvider } from './contexts/WebSocketContext';

function App() {
  return (
    <WebSocketProvider>
      {/* Your app components */}
    </WebSocketProvider>
  );
}

export default App;
```

#### 4. Use the WebSocket Context in Your Components

```tsx
import React from 'react';
import { useWebSocketContext } from '../contexts/WebSocketContext';

function NotificationComponent() {
  const { notifications, isConnected } = useWebSocketContext();
  
  return (
    <div>
      <div className="connection-status">
        {isConnected ? 'Connected' : 'Disconnected'}
      </div>
      
      <h3>Notifications ({notifications.length})</h3>
      
      <ul className="notification-list">
        {notifications.map((notification, index) => (
          <li key={index} className="notification-item">
            <div className="notification-title">{notification.title}</div>
            <div className="notification-message">{notification.message}</div>
            <div className="notification-time">
              {new Date(notification.created_at).toLocaleString()}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default NotificationComponent;
```

## WebSocket Event Types

The backend emits the following event types through the WebSocket connection:

### Order Events

| Event Type | Description | Payload Example |
|------------|-------------|-----------------|
| `order_created` | Triggered when a new order is created | `{ id: "uuid", order_number: "ORD-12345", customer_name: "John Doe", total_amount: 150.00, created_at: "2023-06-15T14:30:00Z" }` |
| `order_status_changed` | Triggered when an order's status changes | `{ id: "uuid", order_number: "ORD-12345", previous_status: "pending", new_status: "processing", updated_at: "2023-06-15T15:30:00Z" }` |
| `order_payment_received` | Triggered when payment is received for an order | `{ id: "uuid", order_number: "ORD-12345", amount: 150.00, payment_method: "credit_card", transaction_id: "tx_123456", received_at: "2023-06-15T15:35:00Z" }` |
| `order_shipped` | Triggered when an order is shipped | `{ id: "uuid", order_number: "ORD-12345", tracking_number: "TRK123456789", carrier: "DHL", shipped_at: "2023-06-16T10:00:00Z" }` |
| `order_delivered` | Triggered when an order is marked as delivered | `{ id: "uuid", order_number: "ORD-12345", delivered_at: "2023-06-18T14:20:00Z" }` |
| `order_cancelled` | Triggered when an order is cancelled | `{ id: "uuid", order_number: "ORD-12345", reason: "Customer request", cancelled_at: "2023-06-15T16:45:00Z" }` |

### Product Events

| Event Type | Description | Payload Example |
|------------|-------------|-----------------|
| `product_created` | Triggered when a new product is created | `{ id: "uuid", name: "Premium T-Shirt", sku: "TS-001", price: 29.99, created_at: "2023-06-14T09:15:00Z" }` |
| `product_updated` | Triggered when a product is updated | `{ id: "uuid", name: "Premium T-Shirt", previous_price: 29.99, new_price: 24.99, updated_at: "2023-06-14T11:30:00Z" }` |
| `product_deleted` | Triggered when a product is deleted | `{ id: "uuid", name: "Premium T-Shirt", deleted_at: "2023-06-14T16:45:00Z" }` |
| `product_price_changed` | Triggered when a product's price changes | `{ id: "uuid", name: "Premium T-Shirt", previous_price: 29.99, new_price: 24.99, updated_at: "2023-06-14T11:30:00Z" }` |

### Inventory Events

| Event Type | Description | Payload Example |
|------------|-------------|-----------------|
| `inventory_updated` | Triggered when inventory levels change | `{ product_id: "uuid", product_name: "Premium T-Shirt", previous_quantity: 100, new_quantity: 95, updated_at: "2023-06-15T14:35:00Z" }` |
| `low_stock` | Triggered when a product's inventory falls below threshold | `{ product_id: "uuid", product_name: "Premium T-Shirt", current_quantity: 5, threshold: 10, alert_time: "2023-06-15T14:35:00Z" }` |
| `out_of_stock` | Triggered when a product goes out of stock | `{ product_id: "uuid", product_name: "Premium T-Shirt", last_sold_at: "2023-06-15T14:35:00Z" }` |
| `back_in_stock` | Triggered when an out-of-stock product is restocked | `{ product_id: "uuid", product_name: "Premium T-Shirt", new_quantity: 50, restocked_at: "2023-06-16T09:00:00Z" }` |

### User Events

| Event Type | Description | Payload Example |
|------------|-------------|-----------------|
| `user_registered` | Triggered when a new user registers | `{ id: "uuid", username: "john_doe", email: "john@example.com", registered_at: "2023-06-10T11:20:00Z" }` |
| `user_login` | Triggered when a user logs in | `{ id: "uuid", username: "john_doe", login_at: "2023-06-15T09:30:00Z", device_info: "Chrome/Windows" }` |
| `password_changed` | Triggered when a user changes their password | `{ id: "uuid", username: "john_doe", changed_at: "2023-06-12T16:45:00Z" }` |
| `user_role_changed` | Triggered when a user's role changes | `{ id: "uuid", username: "john_doe", previous_role: "customer", new_role: "admin", changed_at: "2023-06-13T10:15:00Z" }` |

### System Events

| Event Type | Description | Payload Example |
|------------|-------------|-----------------|
| `system_maintenance` | Triggered for scheduled maintenance notifications | `{ title: "Scheduled Maintenance", message: "The system will be down for maintenance on June 20th from 2-4 AM UTC", scheduled_at: "2023-06-20T02:00:00Z", estimated_duration_minutes: 120 }` |
| `error_alert` | Triggered when a system error occurs | `{ error_code: "ERR-5001", severity: "high", message: "Database connection issue detected", occurred_at: "2023-06-15T18:30:00Z" }` |

## Message Format

All WebSocket messages follow this standard format:

```json
{
  "type": "notification",
  "payload": {
    "id": "uuid",
    "event_type": "order_created",
    "title": "New Order",
    "message": "Order #ORD-12345 has been created",
    "metadata": {
      // Event-specific data as shown in the tables above
    },
    "created_at": "2023-06-15T14:30:00Z",
    "read": false,
    "priority": "normal" // "low", "normal", "high", "urgent"
  }
}
```

## Troubleshooting WebSocket Connections

### Common Connection Issues and Solutions

| Issue | Symptoms | Solution |
|-------|----------|----------|
| Incorrect WebSocket URL | Connection fails immediately | Double-check the URL, ensure it starts with `ws://` and points to your backend server |
| Missing API prefix | 404 errors in console | Ensure the URL includes the required path prefix `/api/ws` |
| Invalid or expired token | 401 Unauthorized errors | Get a fresh token by logging in again, check token expiration time |
| Token format issues | Authentication fails | Ensure the token is sent exactly as received from backend, check for encoding issues |
| CORS issues | Blocked by browser security | Configure the backend to allow WebSocket connections from your frontend origin |
| Server not running WebSocket | Connection attempts fail | Verify the backend WebSocket service is running and properly configured |

### Testing with Command Line Tools

To verify that your WebSocket endpoint is working correctly, use a command-line tool like `websocat`:

```bash
# Install websocat
npm install -g websocat

# Test connection with your token
websocat "ws://localhost:3000/api/ws?token=YOUR_TOKEN_HERE"
```

### WebSocket Protocol for HTTPS

If you're using HTTPS for your frontend, use WSS instead of WS:

```typescript
const wsUrl = import.meta.env.VITE_WS_URL || 
  (window.location.protocol === 'https:' ? 'wss://' : 'ws://') + 
  window.location.host + '/api/ws';
```

## Best Practices

1. **Error Handling**: Always implement proper error handling for WebSocket connections.
2. **Reconnection Logic**: Implement automatic reconnection with exponential backoff.
3. **Notification Storage**: Consider storing notifications locally (localStorage/IndexedDB) for offline access.
4. **Notification Grouping**: Group similar notifications to avoid overwhelming the user.
5. **Read Status Tracking**: Implement a mechanism to mark notifications as read and sync with the server.
6. **Notification Expiry**: Consider implementing an expiry policy for older notifications.

## UI Implementation Recommendations

1. **Notification Badge**: Display an unread count badge on your notification icon.
2. **Toast Notifications**: Show toast notifications for high-priority events.
3. **Notification Center**: Implement a dropdown or sidebar notification center.
4. **Sound Alerts**: Add optional sound alerts for important notifications.
5. **Desktop Notifications**: Use the Web Notifications API for desktop notifications.
6. **Filtering Options**: Allow users to filter notifications by type or priority.

## Security Considerations

1. Always authenticate WebSocket connections using JWT or similar token-based authentication.
2. Validate all incoming messages on the client side before processing.
3. Implement rate limiting to prevent flooding of notifications.
4. Ensure sensitive information is not exposed in notification payloads. 