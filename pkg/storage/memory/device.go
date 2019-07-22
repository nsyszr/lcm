package memory

import (
	"sync"
	"time"

	"github.com/nsyszr/lcm/pkg/model"
	"github.com/nsyszr/lcm/pkg/storage"
)

type deviceStore struct {
	store  map[int32]model.Device
	nextID int32
	sync.RWMutex
}

func newDeviceStore() *deviceStore {
	return &deviceStore{
		store:  make(map[int32]model.Device),
		nextID: 1,
	}
}

func (s *deviceStore) FetchAll() (models map[int32]model.Device, err error) {
	s.RLock()
	defer s.RUnlock()
	models = make(map[int32]model.Device, len(s.store))

	for id, m := range s.store {
		models[id] = m
	}

	return models, nil
}

func (s *deviceStore) FindByID(id int32) (*model.Device, error) {
	s.RLock()
	defer s.RUnlock()
	if m, ok := s.store[id]; ok {
		return &m, nil
	}

	return nil, storage.ErrNotFound
}

func (s *deviceStore) FindByNamespaceAndDeviceID(namespace, deviceID string) (*model.Device, error) {
	s.RLock()
	defer s.RUnlock()

	for _, m := range s.store {
		if m.Namespace == namespace && m.DeviceID == deviceID {
			return &m, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (s *deviceStore) Create(m *model.Device) error {
	s.Lock()
	defer s.Unlock()

	m.ID = s.getNextID()

	// Set default values
	if m.SessionTimeout == 0 {
		m.SessionTimeout = 120
	}
	if m.PingInterval == 0 {
		m.PingInterval = 104
	}
	if m.PongTimeout == 0 {
		m.PongTimeout = 16
	}
	if m.EventsTopic == "" {
		m.EventsTopic = "deviceevent"
	}

	// TODO(DGL) Add validation, eg. PingInterval not greater than SessionTimeout

	m.CreatedAt = time.Now().Round(time.Second).UTC()
	m.UpdatedAt = time.Now().Round(time.Second).UTC()

	s.store[m.ID] = *m

	return nil
}

func (s *deviceStore) Delete(id int32) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.store[id]
	if !ok {
		return storage.ErrNotFound
	}

	delete(s.store, id)

	return nil
}

func (s *deviceStore) getNextID() int32 {
	id := s.nextID
	s.nextID++
	return id
}
