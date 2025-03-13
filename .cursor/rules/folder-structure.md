Folder structure and architectural note
project-root/
├── cmd/
│   └── app/main.go   # Entry point
│
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── product_handler.go
│   │   │   ├── order_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── notification_handler.go
│   │   │   ├── auth_handler.go
│   │   │   ├── websocket_handler.go // Not sure if this should locate here
│   │
│   ├── services/
│   │   ├── product_service.go
│   │   ├── order_service.go
│   │   ├── user_service.go
│   │   ├── notification_service.go
│   │   ├── auth_service.go
│   │   ├── websocket_service.go // Not sure if this should locate here
│   │
│   ├── repositories/
│   │   ├── product_repository.go
│   │   ├── order_repository.go
│   │   ├── user_repository.go
│   │   ├── notification_repository.go
│   │
│   ├── middleware/
│   │   ├── jwt.go
│   │   ├── guard.go
│   │   ├── logger.go
│   │
│   ├── websocket/
│   │   ├── websocket_config.go   # Logic về WebSocket cho business của bạn
│   │   ├── websocket_events.go   # Event cụ thể (notification, update order...)
│   │
│   ├── uploads/
│   │   ├── upload_config.go      # Config upload cho business này
│   │   ├── upload_events.go      # Gán ảnh cho sản phẩm, validation logic
│
│   ├── utils/
│   │   ├── response.go
│   │   ├── pagination.go
│   │   ├── logger.go
│   │
├── pkg/
│   ├── websocket/
│   │   ├── hub.go                # Hub quản lý client WebSocket
│   │   ├── client.go             # Quản lý từng kết nối
│   │   ├── websocket.go          # Khởi tạo WebSocket thuần túy
│   │
│   ├── upload/
│   │   ├── storage.go            # Xử lý lưu file (local/cloud)
│   │   ├── validator.go          # Validate file upload
│   │
│   ├── config/
│   │   ├── config.go             # Viper Config
│   │
│   ├── database/
│   │   ├── db.go                 # DB Connection
│   │
│   ├── jwt/
│   │   ├── jwt.go                # JWT Helper
│
├── scripts/
│   └── migrate.sh
│
├── migrations/
│   └── 2024xxx_init_schema.sql
│
├── docs/ (Swagger generated)
│   └── swagger.json
│
├── Dockerfile
├── docker-compose.yml
├── .env
├── go.mod
├── go.sum
└── main.go


