package orchestration

import (
	"sync"
	"time"
)

type Session struct {
	AuthSessionID       string
	ClientID            string
	UserRef             string
	Mode                string
	DeviceAuthSessionID string
	DeviceChallengeID   string
	SimAuthSessionID    string
	Status              string
	CreatedAt           time.Time
	ExpiresAt           time.Time
	CompletedAt         *time.Time
}

type Store struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

func NewStore() *Store {
	s := &Store{
		sessions: make(map[string]*Session),
	}
	go s.cleanupLoop()
	return s
}

func (s *Store) Set(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clone := *session
	s.sessions[clone.AuthSessionID] = &clone
}

func (s *Store) Get(id string) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	clone := *sess
	return &clone, true
}

func (s *Store) Update(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clone := *session
	s.sessions[clone.AuthSessionID] = &clone
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.cleanup()
	}
}

func (s *Store) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}
