package libserver

import "sync"

type ServerData struct {
	data           map[string]any
	sessionManager SessionManager
	mu             *sync.RWMutex
}

func NewServerData() *ServerData {
	return &ServerData{
		data: make(map[string]any),
		mu:   &sync.RWMutex{},
	}
}

func (s *ServerData) Set(key string, value any) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

func (s *ServerData) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *ServerData) Delete(key string) {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}

func (s *ServerData) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

func (s *ServerData) Clear() {
	s.mu.Lock()
	s.data = make(map[string]any)
	s.mu.Unlock()
}

func (s *ServerData) SetSessionManager(sessionManager SessionManager) {
	s.sessionManager = sessionManager
}

func (s *ServerData) GetSessionManager() SessionManager {
	return s.sessionManager
}
