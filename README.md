# LibServer

LibServer is a lightweight and flexible Go library designed to simplify web server creation. It natively integrates session management and a thread-safe data sharing system, allowing developers to focus on their application's business logic.

## Features

*   **HTTP/HTTPS Server**: Simple configuration to start a secure or non-secure web server.
*   **Automatic Session Management**:
    *   Transparent session management via cookies.
    *   Unique identification by UUID.
    *   Automatic session expiration (default 1 hour).
    *   Thread-safe in-memory storage.
*   **Shared Server Data**:
    *   Global key-value storage (thread-safe) accessible from all handlers.
    *   Automatic injection into the request context.
*   **Extensible Architecture**:
    *   Clear interfaces (`Session`, `SessionManager`) allowing implementation of custom storage strategies (Redis, database, etc.).
*   **Contextual Integration**: Sessions and server data are automatically injected into each HTTP request's context (`context.Context`).

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

LibServer automatically manages session creation and retrieval. You can access the current session via the request context.

```go
server.AddHandlerFunc("/session", func(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the context
	// The key is the application name defined during server creation
	ctx := r.Context()
	session := ctx.Value("MyApp").(libserver.Session)

	// Write to the session
	session.Set("username", "JohnDoe")

	// Read from the session
	if username := session.Get("username"); username != nil {
		fmt.Fprintf(w, "Hello, %s!", username)
	}
})
```

### 4. Shared Server Data (Global State)

`ServerData` allows sharing information between different requests in a thread-safe manner.

```go
server.AddHandlerFunc("/counter", func(w http.ResponseWriter, r *http.Request) {
	// Accessing server data
	// The recommended way is to access via the server object if you have the reference,
	// or via dependency injection.

	// Conceptual example if you have access to the `server` object:
	data := server.GetServerData()

	currentCount := 0
	if val := data.Get("global_visits"); val != nil {
		currentCount = val.(int)
	}

	data.Set("global_visits", currentCount + 1)

	fmt.Fprintf(w, "Global visits: %d", currentCount + 1)
})
```

> **Note**: In the current implementation, `ServerData` is injected into the context but the `serverDataKey` key is not exported. It is therefore recommended to use `server.GetServerData()` or pass the server instance to your handlers.

### 5. Using a Custom Session Manager

You can replace the default session manager (in-memory) with your own implementation (for example to use Redis).

```go
type MyRedisSessionManager struct {
    // ... your implementation
}

// Implement the SessionManager interface...

func main() {
    server := libserver.NewWebServer("MyApp", "localhost", 8080)

    myManager := &MyRedisSessionManager{}
    server.SetSessionManager(myManager)

    server.Start()
}
```

## Architecture

*   **WebServer**: The main entry point. It configures the `http.ServeMux` router, manages the HTTP server lifecycle, and orchestrates dependency injection (sessions, data) into requests.
*   **SessionManager (Interface)**: Defines how sessions are created, retrieved, and deleted.
    *   `DefaultSessionManager`: Default implementation that stores sessions in memory (`map`) and cleans up expired sessions every hour.
*   **Session (Interface)**: Defines operations on a session (Get, Set, Delete, etc.).
    *   `DefaultSession`: Default implementation with expiration after one hour.
*   **ServerData**: A thread-safe structure (`sync.RWMutex`) to store global application data.

## Contribution

Contributions are welcome! Feel free to open an issue or a pull request.

## License

This project is licensed under the [MIT](LICENSE) license.
