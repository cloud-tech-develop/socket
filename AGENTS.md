# Agent Guidelines for socket-go

This is a Go WebSocket chat server using Echo framework and Gorilla WebSocket.

## Build & Run Commands

```bash
# Build the application
go build -o socket-go .

# Run the server
go run main.go

# Run with environment variable
PORT=9000 go run main.go
```

## Docker Commands

```bash
# Build Docker image
docker build -t socket:latest .

# Run container
docker run -d -p 8888:8888 --name socket-app socket:latest

# Run with Docker Compose
docker compose up --build -d
```

## Testing

There are currently no test files in this project. When tests are added:

```bash
# Run all tests
go test ./...

# Run tests in specific package
go test ./internal/app/chat/...

# Run single test
go test -run TestName ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

## Linting & Code Quality

```bash
# Format code (mandatory before commits)
go fmt ./...

# Run go vet
go vet ./...

# Check for unused imports
goimports -l .

# Full check
go build -o /dev/null ./...
```

## Code Style Guidelines

### Imports

- Use standard Go import grouping: standard library first, then third-party
- Use import aliases only when necessary (e.g., `echov4 "github.com/labstack/echo/v4"`)
- Run `goimports` or `go fmt` before committing

### Formatting

- Use `go fmt` for automatic formatting
- Maximum line length: 120 characters (soft limit)
- Use tabs for indentation, not spaces

### Types & Declarations

```go
// Variable declarations - use short variable declarations where possible
port := utils.GetPort()

// Constants - use CamelCase for exported, camelCase for unexported
const MaxClients = 100
const maxMessageSize = 4096

// Interfaces - name with Reader, Writer, Handler suffix when appropriate
type Upgrader interface {
    Upgrade(w http.ResponseWriter, r *http.Request, h http.Header) (*Conn, error)
}

// Structs - use meaningful field names, group related fields
type Client struct {
    ID         string
    Connection *websocket.Conn
    Room       *Room
}
```

### Naming Conventions

- **Variables/Functions**: camelCase (e.g., `handleWebSocket`, `clientID`)
- **Constants**: CamelCase for exported, camelCase for unexported
- **Types/Interfaces**: PascalCase (e.g., `Client`, `Room`, `Message`)
- **Packages**: lowercase, short, no underscores (e.g., `utils`, `chat`)
- **Files**: lowercase with underscores for multi-word names (e.g., `chat_test.go`)

### Error Handling

- Always check errors explicitly - do not ignore with `_`
- Return meaningful error messages
- Use `log` or structured logging for debugging
- Wrap errors with context where helpful: `fmt.Errorf("failed to join room: %w", err)`

```go
// Good
if err := joinRoom(client, roomID); err != nil {
    log.Printf("Error al unirse a la sala: %s", err)
    conn.Close()
    return echo.NewHTTPError(500, "Error al unirse a la sala")
}

// Avoid
_ = joinRoom(client, roomID)
```

### Concurrency

- Use `sync.Mutex` for protecting shared state
- Always `defer mu.Unlock()` after `mu.Lock()`
- Use `defer` for cleanup operations (closing connections, leaving rooms)

```go
mu.Lock()
defer mu.Unlock()
// operations on shared state
```

### JSON Handling

- Use struct tags for JSON serialization
- Use camelCase for JSON field names in tags (Go convention)

```go
type Message struct {
    Sender   int64     `json:"sender"`
    Content  string    `json:"content"`
    DateTime time.Time `json:"dateTime"`
}
```

### WebSocket Patterns

- Use `websocket.Upgrader` for connection upgrades
- Check for close errors properly:
  ```go
  if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
      // Handle graceful disconnect
  }
  ```
- Always close connections in `defer`
- Handle write errors and remove disconnected clients

## Project Structure

```
socket-go/
├── main.go                 # Application entry point
├── internal/
│   ├── app/
│   │   └── chat/           # Chat logic, WebSocket handling
│   └── utils/              # Utility functions (port config)
├── pkg/
│   └── socket_cors/        # WebSocket CORS configuration
├── go.mod                  # Module dependencies
├── Dockerfile              # Container image
└── docker-compose.yml      # Container orchestration
```

## Key Dependencies

- `github.com/labstack/echo/v4` - HTTP router and middleware
- `github.com/gorilla/websocket` - WebSocket implementation
- `github.com/google/uuid` - Unique ID generation

## Environment Variables

- `PORT` - Server port (default: 8888)

## Common Tasks

### Adding a New Endpoint

1. Add handler function in appropriate package (likely `internal/app/chat/`)
2. Register route in `routes.go` using Echo's routing methods
3. Test with curl or browser

### Adding a New WebSocket Feature

1. Update `Client`, `Room`, or `Message` structs in `common.go`
2. Implement handler logic in `chat.go`
3. Add appropriate error handling and logging

### Modifying CORS Configuration

Edit `pkg/socket_cors/socket_cors.go` to adjust WebSocket upgrader settings.

## Notes

- This project is in Spanish - error messages and logs are in Spanish
- No tests exist yet - consider adding unit tests for chat logic
- The server uses mutex-based concurrency, not channels