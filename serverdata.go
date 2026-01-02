package libserver

import "sync"

// ServerData is a thread-safe key-value store for global application data
type ServerData struct {
	data           map[string]any
	sessionManager SessionManager
	mu             *sync.RWMutex
}

// NewServerData creates a new ServerData instance
func NewServerData() *ServerData {
	return &ServerData{
		data: make(map[string]any),
		mu:   &sync.RWMutex{},
	}
}

// Set stores a value with the given key
func (s *ServerData) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a value by its key, returns nil if not found
func (s *ServerData) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

// Delete removes a value by its key
func (s *ServerData) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Has checks if a key exists
func (s *ServerData) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

// Clear removes all data
func (s *ServerData) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]any)
}

// SetSessionManager sets the session manager
func (s *ServerData) SetSessionManager(sessionManager SessionManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionManager = sessionManager
}

// GetSessionManager returns the session manager
func (s *ServerData) GetSessionManager() SessionManager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionManager
}

// Keys returns all keys in the store
func (s *ServerData) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Len returns the number of items in the store
func (s *ServerData) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}
