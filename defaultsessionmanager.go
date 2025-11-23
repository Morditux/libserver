package libserver

import (
	"sync"
	"time"
)

type DefaultSessionManager struct {
	data map[string]Session
	mu   *sync.RWMutex
}

func NewDefaultSessionManager() *DefaultSessionManager {
	manager := &DefaultSessionManager{
		data: make(map[string]Session),
		mu:   &sync.RWMutex{},
	}
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			manager.Clear()
		}
	}()
	return manager
}

func (s *DefaultSessionManager) CreateSession() Session {
	return NewDefaultSession()
}

func (s *DefaultSessionManager) GetSession(id string) Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

func (s *DefaultSessionManager) DeleteSession(id string) {
	s.mu.Lock()
	delete(s.data, id)
	s.mu.Unlock()
}

func (s *DefaultSessionManager) HasSession(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[id]
	return ok
}

func (s *DefaultSessionManager) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Remove expired sessions
	for id, session := range s.data {
		if session.IsExpired() {
			delete(s.data, id)
		}
	}
}
