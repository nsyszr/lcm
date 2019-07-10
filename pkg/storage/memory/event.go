package memory

import (
	"sync"
	"time"

	"github.com/nsyszr/lcm/pkg/model"
	"github.com/nsyszr/lcm/pkg/storage"
)

type eventStore struct {
	store  map[int32]model.Event
	nextID int32
	sync.RWMutex
}

func newEventStore() *eventStore {
	return &eventStore{
		store:  make(map[int32]model.Event),
		nextID: 1,
	}
}

func (s *eventStore) FetchAll() (models map[int32]model.Event, err error) {
	s.RLock()
	defer s.RUnlock()
	models = make(map[int32]model.Event, len(s.store))

	for id, m := range s.store {
		models[id] = m
	}

	return models, nil
}

func (s *eventStore) FindByID(id int32) (*model.Event, error) {
	s.RLock()
	defer s.RUnlock()
	if m, ok := s.store[id]; ok {
		return &m, nil
	}

	return nil, storage.ErrNotFound
}

func (s *eventStore) Create(m *model.Event) error {
	s.Lock()
	defer s.Unlock()

	m.ID = s.getNextID()
	m.CreatedAt = time.Now().Round(time.Second).UTC()
	m.UpdatedAt = time.Now().Round(time.Second).UTC()

	s.store[m.ID] = *m

	return nil
}

func (s *eventStore) getNextID() int32 {
	id := s.nextID
	s.nextID++
	return id
}
