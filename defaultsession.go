package libserver

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type DefaultSession struct {
	data      map[string]any
	createdAt time.Time
	mu        *sync.RWMutex
	id        string
}

func NewDefaultSession() *DefaultSession {
	return &DefaultSession{
		data:      make(map[string]any),
		createdAt: time.Now(),
		mu:        &sync.RWMutex{},
		id:        uuid.New().String(),
	}
}

func (s *DefaultSession) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *DefaultSession) Set(key string, value any) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

func (s *DefaultSession) Delete(key string) {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}

func (s *DefaultSession) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

func (s *DefaultSession) Clear() {
	s.mu.Lock()
	s.data = make(map[string]any)
	s.mu.Unlock()
}

func (s *DefaultSession) IsExpired() bool {
	return time.Since(s.createdAt) > time.Hour
}

func (s *DefaultSession) Id() string {
	return s.id
}
