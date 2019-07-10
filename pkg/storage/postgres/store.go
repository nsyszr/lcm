package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/nsyszr/lcm/pkg/storage"
)

// store contains all PostgreSQL based sub-stores for managing the models
type store struct {
	sessions *sessionStore
	events   *eventStore
	devices  *deviceStore
}

// NewStore creates a new PostgreSQL based Storage interface
func NewStore(db *sqlx.DB) storage.Interface {
	return &store{
		sessions: newSessionStore(db),
		events:   newEventStore(db),
		devices:  newDeviceStore(db),
	}
}

// Sessions returns a sub-store for managing the Session model
func (s *store) Sessions() storage.SessionStore {
	return s.sessions
}

// Events returns a sub-store for managing the Event model
func (s *store) Events() storage.EventStore {
	return s.events
}

// Devices returns a sub-store for managing the Event model
func (s *store) Devices() storage.DeviceStore {
	return s.devices
}
