package store

import (
	"log"
	"sync"
	"time"

	"device_only/internal/sim/model"
)

// SessionStore is a mutex-protected in-memory store for authentication sessions.
// All public methods return copies of sessions so callers cannot accidentally
// mutate stored state without going through Update.
type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]*model.Session
}

// NewSessionStore creates a SessionStore and starts the background cleanup goroutine.
func NewSessionStore() *SessionStore {
	s := &SessionStore{
		sessions: make(map[string]*model.Session),
	}
	go s.cleanupLoop()
	return s
}

// Set stores a new session. A defensive copy is made so the caller may reuse the
// original struct safely.
func (s *SessionStore) Set(session *model.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clone := *session
	s.sessions[clone.AuthSessionID] = &clone
}

// Get retrieves a session by ID and returns a copy. The second return value is
// false when the ID is not found.
func (s *SessionStore) Get(id string) (*model.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	clone := *sess
	return &clone, true
}

// Update overwrites the stored session for session.AuthSessionID with a copy of
// the supplied value.
func (s *SessionStore) Update(session *model.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clone := *session
	s.sessions[clone.AuthSessionID] = &clone
}

// cleanupLoop runs every 60 seconds and removes sessions whose TTL has elapsed.
func (s *SessionStore) cleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.cleanup()
	}
}

func (s *SessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	removed := 0
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
			removed++
		}
	}
	if removed > 0 {
		log.Printf("[INFO] Session cleanup: removed %d expired session(s)", removed)
	}
}
