package memory

import "github.com/nsyszr/lcm/pkg/storage"

// Store contains all memory-based sub-stores for managing the persistent models
type store struct {
	sessions *sessionStore
	events   *eventStore
}

// NewStore creates a new memory-based Storage interface
func NewStore() storage.Interface {
	sessionStore := newSessionStore()
	eventStore := newEventStore()

	return &store{
		sessions: sessionStore,
		events:   eventStore,
	}
}

// Sessions returns a sub-store for managing the Session model
func (s *store) Sessions() storage.SessionStore {
	return s.sessions
}

// Events returns a sub-store for managing the license spec model
func (s *store) Events() storage.EventStore {
	return s.events
}
