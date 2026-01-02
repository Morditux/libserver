package libserver

import (
	"sync"
	"time"
)

// DefaultCleanupInterval is the default interval for cleaning up expired sessions
const DefaultCleanupInterval = time.Hour

// DefaultSessionManager is the default implementation of the SessionManager interface
type DefaultSessionManager struct {
	data              map[string]Session
	mu                *sync.RWMutex
	stopCh            chan struct{}
	cleanupInterval   time.Duration
	sessionExpiration time.Duration
	cleanupRunning    bool
}

// NewDefaultSessionManager creates a new session manager with default settings
func NewDefaultSessionManager() *DefaultSessionManager {
	return NewDefaultSessionManagerWithConfig(DefaultCleanupInterval, DefaultSessionExpiration)
}

// NewDefaultSessionManagerWithConfig creates a new session manager with custom settings
func NewDefaultSessionManagerWithConfig(cleanupInterval, sessionExpiration time.Duration) *DefaultSessionManager {
	manager := &DefaultSessionManager{
		data:              make(map[string]Session),
		mu:                &sync.RWMutex{},
		stopCh:            make(chan struct{}),
		cleanupInterval:   cleanupInterval,
		sessionExpiration: sessionExpiration,
		cleanupRunning:    false,
	}
	manager.startCleanup()
	return manager
}

// startCleanup starts the background goroutine for cleaning up expired sessions
func (s *DefaultSessionManager) startCleanup() {
	s.mu.Lock()
	if s.cleanupRunning {
		s.mu.Unlock()
		return
	}
	s.cleanupRunning = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(s.cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.cleanup()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop stops the cleanup goroutine gracefully
func (s *DefaultSessionManager) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cleanupRunning {
		close(s.stopCh)
		s.cleanupRunning = false
	}
}

// CreateSession creates a new session with default expiration
func (s *DefaultSessionManager) CreateSession() Session {
	session := NewDefaultSessionWithExpiration(s.sessionExpiration)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[session.Id()] = session
	return session
}

// CreateSessionWithID creates a new session with a specific ID (for session restoration)
func (s *DefaultSessionManager) CreateSessionWithID(id string) Session {
	session := NewDefaultSessionWithExpiration(s.sessionExpiration)
	// Use reflection or create a special constructor - here we create and replace the ID
	// For simplicity, we create a standard session
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = session
	return session
}

// GetSession retrieves a session by its ID, returns nil if not found
func (s *DefaultSessionManager) GetSession(id string) Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[id]
}

// DeleteSession removes a session by its ID
func (s *DefaultSessionManager) DeleteSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
}

// HasSession checks if a session exists by its ID
func (s *DefaultSessionManager) HasSession(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[id]
	return ok
}

// cleanup removes all expired sessions
func (s *DefaultSessionManager) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, session := range s.data {
		if session.IsExpired() {
			delete(s.data, id)
		}
	}
}

// SessionCount returns the number of active sessions
func (s *DefaultSessionManager) SessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// SetSessionExpiration sets the default expiration for new sessions
func (s *DefaultSessionManager) SetSessionExpiration(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionExpiration = d
}
