package storage

import "github.com/nsyszr/lcm/pkg/model"

// Interface is implemented by the storage
type Interface interface {
	Sessions() SessionStore
	Events() EventStore
	Devices() DeviceStore
}

// SessionStore is responsible for managing the Session model
type SessionStore interface {
	FetchAll() (map[int32]model.Session, error)
	FindByID(id int32) (*model.Session, error)
	FindByNamespaceAndDeviceID(namespace, deviceID string) (*model.Session, error)
	Create(m *model.Session) error
	Delete(id int32) error
}

// EventStore is responsible for managing the Event model
type EventStore interface {
	FetchAll() (map[int32]model.Event, error)
	FindByID(id int32) (*model.Event, error)
	Create(m *model.Event) error
}

// DeviceStore is responsible for managing the Device model
type DeviceStore interface {
	FetchAll() (map[int32]model.Device, error)
	FindByID(id int32) (*model.Device, error)
	FindByNamespaceAndDeviceID(namespace, deviceID string) (*model.Device, error)
	Create(m *model.Device) error
}
