package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/nsyszr/lcm/pkg/model"
)

type sessionStore struct {
	store  map[int32]model.Session
	nextID int32
	sync.RWMutex
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		store:  make(map[int32]model.Session),
		nextID: 1,
	}
}

func (s *sessionStore) FetchAll() (models map[int32]model.Session, err error) {
	s.RLock()
	defer s.RUnlock()
	models = make(map[int32]model.Session, len(s.store))

	for id, m := range s.store {
		models[id] = m
	}

	return models, nil
}

func (s *sessionStore) FindByID(id int32) (*model.Session, error) {
	s.RLock()
	defer s.RUnlock()
	if m, ok := s.store[id]; ok {
		return &m, nil
	}

	return nil, fmt.Errorf("not found")
}

func (s *sessionStore) FindByNamespaceAndDeviceID(namespace, deviceID string) (*model.Session, error) {
	s.RLock()
	defer s.RUnlock()

	for _, m := range s.store {
		if m.Namespace == namespace && m.DeviceID == deviceID {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (s *sessionStore) Create(m *model.Session) error {
	s.Lock()
	defer s.Unlock()

	m.ID = s.getNextID()
	m.CreatedAt = time.Now().Round(time.Second).UTC()
	m.UpdatedAt = time.Now().Round(time.Second).UTC()

	s.store[m.ID] = *m

	return nil
}

func (s *sessionStore) Delete(id int32) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.store[id]
	if !ok {
		return fmt.Errorf("not found")
	}

	delete(s.store, id)

	return nil
}

func (s *sessionStore) getNextID() int32 {
	id := s.nextID
	s.nextID++
	return id
}
