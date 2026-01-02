package libserver

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// DefaultSessionExpiration is the default session expiration duration
const DefaultSessionExpiration = time.Hour

// DefaultSession is the default implementation of the Session interface
type DefaultSession struct {
	data               map[string]any
	createdAt          time.Time
	lastAccessedAt     time.Time
	mu                 *sync.RWMutex
	id                 string
	expirationDuration time.Duration
}

// NewDefaultSession creates a new session with default expiration (1 hour)
func NewDefaultSession() *DefaultSession {
	return NewDefaultSessionWithExpiration(DefaultSessionExpiration)
}

// NewDefaultSessionWithExpiration creates a new session with a custom expiration duration
func NewDefaultSessionWithExpiration(expiration time.Duration) *DefaultSession {
	now := time.Now()
	return &DefaultSession{
		data:               make(map[string]any),
		createdAt:          now,
		lastAccessedAt:     now,
		mu:                 &sync.RWMutex{},
		id:                 uuid.New().String(),
		expirationDuration: expiration,
	}
}

// Get retrieves a value from the session
func (s *DefaultSession) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

// Set stores a value in the session
func (s *DefaultSession) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete removes a value from the session
func (s *DefaultSession) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Has checks if a key exists in the session
func (s *DefaultSession) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

// Clear removes all data from the session
func (s *DefaultSession) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]any)
}

// IsExpired returns true if the session has expired based on last access time
func (s *DefaultSession) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.lastAccessedAt) > s.expirationDuration
}

// Id returns the session's unique identifier
func (s *DefaultSession) Id() string {
	return s.id
}

// CreatedAt returns the time when the session was created
func (s *DefaultSession) CreatedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.createdAt
}

// LastAccessedAt returns the time when the session was last accessed
func (s *DefaultSession) LastAccessedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastAccessedAt
}

// Update refreshes the session's last access time
func (s *DefaultSession) Update() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastAccessedAt = time.Now()
}

// ExpirationDuration returns the session's expiration duration
func (s *DefaultSession) ExpirationDuration() time.Duration {
	return s.expirationDuration
}

// SetExpirationDuration sets the session's expiration duration
func (s *DefaultSession) SetExpirationDuration(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expirationDuration = d
}
