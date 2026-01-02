# LibServer

LibServer is a lightweight and flexible Go library designed to simplify web server creation. It natively integrates session management and a thread-safe data sharing system, allowing developers to focus on their application's business logic.

## Features

*   **HTTP/HTTPS Server**: Simple configuration to start a secure or non-secure web server.
*   **Automatic Session Management**:
    *   Transparent session management via cookies.
    *   Unique identification by UUID.
    *   Configurable session expiration (default 1 hour, based on last access time).
    *   Thread-safe in-memory storage.
    *   Automatic cleanup of expired sessions.
*   **Shared Server Data**:
    *   Global key-value storage (thread-safe) accessible from all handlers.
    *   Automatic injection into the request context.
    *   Helper functions for easy context retrieval.
*   **Extensible Architecture**:
    *   Clear interfaces (`Session`, `SessionManager`) allowing implementation of custom storage strategies (Redis, database, etc.).
*   **Contextual Integration**: Sessions and server data are automatically injected into each HTTP request's context (`context.Context`).
*   **Graceful Shutdown**: Proper cleanup of goroutines and resources when stopping the server.

## Installation

To install LibServer in your Go project, use the following command:

```bash
go get github.com/Morditux/libserver
```

## Usage

### 1. Creating and Starting a Simple Server

Here is a minimal example to start an HTTP server on port 8080.

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/Morditux/libserver"
)

func main() {
	// Create a new WebServer instance
	// Parameters: Application Name (used for the session cookie), Address, Port
	server := libserver.NewWebServer("MyApp", "localhost", 8080)

	// Add a route
	server.AddHandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to my server!")
	})

	// Start the server
	fmt.Println("Server started on http://localhost:8080")
	if err := server.Start(); err != nil {
		panic(err)
	}
}
```

### 2. Using HTTPS

To enable HTTPS, you must provide the paths to your certificate and private key.

```go
server := libserver.NewWebServer("MyAppSecure", "localhost", 8443)
server.EnableHTTPS("cert.pem", "key.pem")
server.Start()
```

### 3. Session Management

LibServer automatically manages session creation and retrieval. You can access the current session via helper functions or directly from the context.

#### Using Helper Functions (Recommended)

```go
server.AddHandlerFunc("/session", func(w http.ResponseWriter, r *http.Request) {
	// Get session using the helper function
	session := libserver.GetSessionFromContext(r.Context(), "MyApp")
	if session == nil {
		http.Error(w, "No session", http.StatusInternalServerError)
		return
	}

	// Write to the session
	session.Set("username", "JohnDoe")

	// Read from the session
	if username := session.Get("username"); username != nil {
		fmt.Fprintf(w, "Hello, %s!", username)
	}
})
```

#### Direct Context Access

```go
server.AddHandlerFunc("/session", func(w http.ResponseWriter, r *http.Request) {
	// The key is the ContextKey type with the application name
	session := r.Context().Value(libserver.ContextKey("MyApp")).(libserver.Session)

	session.Set("username", "JohnDoe")
	if username := session.Get("username"); username != nil {
		fmt.Fprintf(w, "Hello, %s!", username)
	}
})
```

### 4. Shared Server Data (Global State)

`ServerData` allows sharing information between different requests in a thread-safe manner.

#### Using Helper Functions (Recommended)

```go
server.AddHandlerFunc("/counter", func(w http.ResponseWriter, r *http.Request) {
	// Get server data from context
	data := libserver.GetServerDataFromContext(r.Context())
	if data == nil {
		http.Error(w, "No server data", http.StatusInternalServerError)
		return
	}

	currentCount := 0
	if val := data.Get("global_visits"); val != nil {
		currentCount = val.(int)
	}

	data.Set("global_visits", currentCount+1)
	fmt.Fprintf(w, "Global visits: %d", currentCount+1)
})
```

#### Direct Access via Server

```go
server.AddHandlerFunc("/counter", func(w http.ResponseWriter, r *http.Request) {
	data := server.GetServerData()

	currentCount := 0
	if val := data.Get("global_visits"); val != nil {
		currentCount = val.(int)
	}

	data.Set("global_visits", currentCount+1)
	fmt.Fprintf(w, "Global visits: %d", currentCount+1)
})
```

### 5. Configuring Session Expiration

You can customize session expiration when creating the session manager.

```go
import "time"

func main() {
	server := libserver.NewWebServer("MyApp", "localhost", 8080)

	// Create a session manager with custom settings:
	// - Cleanup interval: every 30 minutes
	// - Session expiration: 2 hours of inactivity
	sessionManager := libserver.NewDefaultSessionManagerWithConfig(
		30*time.Minute,  // cleanup interval
		2*time.Hour,     // session expiration
	)
	server.SetSessionManager(sessionManager)

	server.Start()
}
```

### 6. Using a Custom Session Manager

You can replace the default session manager (in-memory) with your own implementation (for example to use Redis).

```go
type MyRedisSessionManager struct {
    // ... your implementation
}

// Implement the SessionManager interface:
// - CreateSession() Session
// - GetSession(id string) Session
// - DeleteSession(id string)
// - HasSession(id string) bool

func main() {
    server := libserver.NewWebServer("MyApp", "localhost", 8080)

    myManager := &MyRedisSessionManager{}
    server.SetSessionManager(myManager)

    server.Start()
}
```

### 7. Graceful Shutdown

The server supports graceful shutdown with proper cleanup of resources.

```go
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/Morditux/libserver"
)

func main() {
	server := libserver.NewWebServer("MyApp", "localhost", 8080)

	server.AddHandlerFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello!")
	})

	// Handle shutdown signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("Shutting down server...")
		if err := server.Stop(); err != nil {
			fmt.Printf("Error stopping server: %v\n", err)
		}
	}()

	fmt.Println("Server started on http://localhost:8080")
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
```

## Architecture

*   **WebServer**: The main entry point. It configures the `http.ServeMux` router, manages the HTTP server lifecycle, and orchestrates dependency injection (sessions, data) into requests.
*   **SessionManager (Interface)**: Defines how sessions are created, retrieved, and deleted.
    *   `DefaultSessionManager`: Default implementation that stores sessions in memory (`map`) and cleans up expired sessions periodically. Supports graceful shutdown via `Stop()`.
*   **Session (Interface)**: Defines operations on a session (Get, Set, Delete, etc.).
    *   `DefaultSession`: Default implementation with configurable expiration based on last access time.
*   **ServerData**: A thread-safe structure (`sync.RWMutex`) to store global application data.

## API Reference

### Context Keys

| Key | Description |
|-----|-------------|
| `libserver.ServerDataKey` | Context key for accessing `*ServerData` |
| `libserver.ContextKey(appName)` | Context key for accessing the `Session` |

### Helper Functions

| Function | Description |
|----------|-------------|
| `GetSessionFromContext(ctx, appName)` | Retrieves the session from context |
| `GetServerDataFromContext(ctx)` | Retrieves the server data from context |

### DefaultSession Methods

| Method | Description |
|--------|-------------|
| `Get(key)` | Retrieves a value |
| `Set(key, value)` | Stores a value |
| `Delete(key)` | Removes a value |
| `Has(key)` | Checks if key exists |
| `Clear()` | Removes all data |
| `IsExpired()` | Checks if session is expired |
| `Update()` | Refreshes last access time |
| `Id()` | Returns session ID |
| `CreatedAt()` | Returns creation time |
| `LastAccessedAt()` | Returns last access time |
| `ExpirationDuration()` | Returns expiration duration |
| `SetExpirationDuration(d)` | Sets expiration duration |

### DefaultSessionManager Methods

| Method | Description |
|--------|-------------|
| `CreateSession()` | Creates a new session |
| `GetSession(id)` | Retrieves a session by ID |
| `DeleteSession(id)` | Deletes a session |
| `HasSession(id)` | Checks if session exists |
| `Stop()` | Stops cleanup goroutine |
| `SessionCount()` | Returns number of active sessions |
| `SetSessionExpiration(d)` | Sets default session expiration |

### ServerData Methods

| Method | Description |
|--------|-------------|
| `Get(key)` | Retrieves a value |
| `Set(key, value)` | Stores a value |
| `Delete(key)` | Removes a value |
| `Has(key)` | Checks if key exists |
| `Clear()` | Removes all data |
| `Keys()` | Returns all keys |
| `Len()` | Returns number of items |

## Contribution

Contributions are welcome! Feel free to open an issue or a pull request.

## License

This project is licensed under the [MIT](LICENSE) license.
