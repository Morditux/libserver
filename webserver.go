package libserver

import (
	"context"
	"fmt"
	"net/http"
)

type WebServer struct {
	address        string
	port           int
	server         *http.Server
	mux            *http.ServeMux
	data           *ServerData
	certFile       string
	keyFile        string
	withHttps      bool
	sessionManager SessionManager
}

func NewWebServer(address string, port int) *WebServer {
	return &WebServer{
		address:   address,
		port:      port,
		server:    &http.Server{Addr: fmt.Sprintf("%s:%d", address, port)},
		mux:       http.NewServeMux(),
		data:      NewServerData(),
		withHttps: false,
	}
}

func (s *WebServer) Start() error {
	// Set default session manager if none is provided
	if s.sessionManager == nil {
		s.sessionManager = NewDefaultSessionManager()
	}
	s.data.SetSessionManager(s.sessionManager)
	// if https is enabled, use ListenAndServeTLS
	if s.withHttps {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	// else use ListenAndServe
	return s.server.ListenAndServe()
}

func (s *WebServer) EnableHTTPS(certFile, keyFile string) {
	s.certFile = certFile
	s.keyFile = keyFile
	s.withHttps = true
}

func (s *WebServer) Stop() error {
	return s.server.Shutdown(context.Background())
}

func (s *WebServer) AddHandlerFunc(pattern string, handler http.HandlerFunc) {
	// Inject server data into handler
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "serverData", s.data)
		handler(w, r.WithContext(ctx))
	})
}

func (s *WebServer) AddHandler(pattern string, handler http.Handler) {
	// Inject server data into handler
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "serverData", s.data)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *WebServer) GetServerData() *ServerData {
	return s.data
}

func (s *WebServer) GetServer() *http.Server {
	return s.server
}

func (s *WebServer) SetSessionManager(sessionManager SessionManager) {
	s.sessionManager = sessionManager
	s.data.SetSessionManager(sessionManager)
}

func (s *WebServer) GetSessionManager() SessionManager {
	return s.sessionManager
}
