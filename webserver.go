package libserver

import (
	"context"
	"fmt"
	"net/http"
)

// ContextKey is the type used for context keys to avoid collisions
type ContextKey string

const (
	// ServerDataKey is the context key for accessing ServerData
	ServerDataKey ContextKey = "serverData"
)

// WebServer is the main HTTP server with integrated session management
type WebServer struct {
	applicationName string
	address         string
	port            int
	server          *http.Server
	mux             *http.ServeMux
	data            *ServerData
	certFile        string
	keyFile         string
	withHttps       bool
	sessionManager  SessionManager
}

// NewWebServer creates a new WebServer instance
func NewWebServer(name, address string, port int) *WebServer {
	return &WebServer{
		address:         address,
		port:            port,
		server:          &http.Server{Addr: fmt.Sprintf("%s:%d", address, port)},
		mux:             http.NewServeMux(),
		data:            NewServerData(),
		withHttps:       false,
		applicationName: name,
	}
}

// Start starts the web server
func (s *WebServer) Start() error {
	// Set default session manager if none is provided
	if s.sessionManager == nil {
		s.sessionManager = NewDefaultSessionManager()
	}
	s.data.SetSessionManager(s.sessionManager)

	// Set the handler
	s.server.Handler = s.mux

	// if https is enabled, use ListenAndServeTLS
	if s.withHttps {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	// else use ListenAndServe
	return s.server.ListenAndServe()
}

// EnableHTTPS enables HTTPS with the provided certificate and key files
func (s *WebServer) EnableHTTPS(certFile, keyFile string) {
	s.certFile = certFile
	s.keyFile = keyFile
	s.withHttps = true
}

// Stop gracefully shuts down the web server
func (s *WebServer) Stop() error {
	// Stop the session manager cleanup goroutine if it's the default one
	if defaultManager, ok := s.sessionManager.(*DefaultSessionManager); ok {
		defaultManager.Stop()
	}
	return s.server.Shutdown(context.Background())
}

// wrapHandler wraps a handler function with session and server data injection
func (s *WebServer) wrapHandler(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session from cookie, if none create one
		session := s.getOrCreateSession(w, r)

		// Update session last access time
		session.Update()

		// Inject server data and session into context
		ctx := context.WithValue(r.Context(), ServerDataKey, s.data)
		ctx = context.WithValue(ctx, ContextKey(s.applicationName), session)

		handler(w, r.WithContext(ctx))
	}
}

// getOrCreateSession retrieves or creates a session for the request
func (s *WebServer) getOrCreateSession(w http.ResponseWriter, r *http.Request) Session {
	sessionCookie, err := r.Cookie(s.applicationName)
	if err == nil {
		// Cookie exists, try to get the session
		if session := s.sessionManager.GetSession(sessionCookie.Value); session != nil && !session.IsExpired() {
			return session
		}
	}

	// Create a new session
	session := s.sessionManager.CreateSession()
	http.SetCookie(w, &http.Cookie{
		Name:     s.applicationName,
		Value:    session.Id(),
		Path:     "/",
		HttpOnly: true,
		Secure:   s.withHttps,
		SameSite: http.SameSiteLaxMode,
	})
	return session
}

// AddHandlerFunc adds a handler function for the given pattern
func (s *WebServer) AddHandlerFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, s.wrapHandler(handler))
}

// AddHandler adds a handler for the given pattern
func (s *WebServer) AddHandler(pattern string, handler http.Handler) {
	s.mux.HandleFunc(pattern, s.wrapHandler(handler.ServeHTTP))
}

// GetServerData returns the server's shared data store
func (s *WebServer) GetServerData() *ServerData {
	return s.data
}

// GetServer returns the underlying http.Server
func (s *WebServer) GetServer() *http.Server {
	return s.server
}

// SetSessionManager sets a custom session manager
func (s *WebServer) SetSessionManager(sessionManager SessionManager) {
	s.sessionManager = sessionManager
	s.data.SetSessionManager(sessionManager)
}

// GetSessionManager returns the current session manager
func (s *WebServer) GetSessionManager() SessionManager {
	return s.sessionManager
}

// GetApplicationName returns the application name
func (s *WebServer) GetApplicationName() string {
	return s.applicationName
}

// GetAddress returns the server address
func (s *WebServer) GetAddress() string {
	return s.address
}

// GetPort returns the server port
func (s *WebServer) GetPort() int {
	return s.port
}

// GetSessionFromContext retrieves the session from a request context
func GetSessionFromContext(ctx context.Context, appName string) Session {
	if session, ok := ctx.Value(ContextKey(appName)).(Session); ok {
		return session
	}
	return nil
}

// GetServerDataFromContext retrieves the ServerData from a request context
func GetServerDataFromContext(ctx context.Context) *ServerData {
	if data, ok := ctx.Value(ServerDataKey).(*ServerData); ok {
		return data
	}
	return nil
}
