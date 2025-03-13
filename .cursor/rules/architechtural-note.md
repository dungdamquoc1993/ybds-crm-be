📌 Overall Architecture
The project follows a three-layer architecture:

Repository Layer: Responsible for interacting with the database via GORM.
Service Layer: Handles business logic and coordinates multiple repositories.
Handler Layer: Receives requests from the client and calls the service for processing.
Additionally, the system includes several supporting modules:

WebSocket: Manages real-time connections.
Upload: Handles file storage.
Middleware: Includes authentication, authorization, and logging.
etc...
📌 Naming Conventions
1️⃣ Repository Layer
Contains CRUD functions that interact directly with the database.
File naming: {model}_repository.go
Method prefixes:
"Find", "Create", "Update", "Delete"
2️⃣ Service Layer
Only uses repositories within its own database cluster.
Uses transactions when multiple data changes are involved.
File naming: {feature}_service.go
3️⃣ Handler Layer (Controller)
Only calls services, does not contain business logic.
File naming: {feature}_handler.go
Method prefixes:
"Handle"
📌 Supporting Modules
✅ WebSocket

pkg/websocket/: Contains core WebSocket logic.
internal/websocket/: Contains business logic specific to this system.
✅ Upload

pkg/upload/: Handles general file storage.
internal/upload/: Contains logic related to storing product images.
📌 Middleware
JWT Middleware (jwt.go)
Authenticates JWT and sets user info in ctx.Locals("user", userData).
RBAC Middleware (guard.go)
Checks permissions based on user.Role.
✅ Database
Uses GORM to interact with PostgreSQL.
📌 General Principles
✅ Handlers must not access repositories directly—only through services.
✅ Services must not access repositories outside their own database cluster.
✅ Transactions must be used for all critical data modifications.
✅ Swagger documentation must be written for all APIs from the start.

This structured translation maintains clarity and follows best practices. Let me know if you'd like any refinements! 🚀
