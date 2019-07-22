package resource

import (
	"fmt"
	"sort"
	"time"

	"github.com/nsyszr/lcm/pkg/model"
)

type DeviceResource struct {
	ID             int32      `json:"id"`
	Namespace      string     `json:"namespace"`
	DeviceID       string     `json:"deviceId"`
	DeviceURI      string     `json:"deviceUri"`
	SessionTimeout int        `json:"sessionTimeout"`
	PingInterval   int        `json:"pingInterval"`
	PongTimeout    int        `json:"pongTimeout"`
	EventsTopic    string     `json:"eventsTopic"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
}

type DeviceListResource struct {
	Members []*DeviceResource `json:"members"`
}

func NewDevice(m *model.Device) (out *DeviceResource) {
	out = &DeviceResource{
		ID:             m.ID,
		Namespace:      m.Namespace,
		DeviceID:       m.DeviceID,
		DeviceURI:      m.DeviceURI,
		SessionTimeout: m.SessionTimeout,
		PingInterval:   m.PingInterval,
		PongTimeout:    m.PongTimeout,
		EventsTopic:    m.EventsTopic,
	}

	if !m.CreatedAt.IsZero() {
		out.CreatedAt = &time.Time{}
		*out.CreatedAt = m.CreatedAt.Round(time.Second)
	}
	if !m.UpdatedAt.IsZero() {
		out.UpdatedAt = &time.Time{}
		*out.UpdatedAt = m.UpdatedAt.Round(time.Second)
	}

	return // out
}

func NewDeviceList(m map[int32]model.Device) (out *DeviceListResource) {
	out = &DeviceListResource{
		Members: make([]*DeviceResource, 0),
	}

	for _, elem := range m {
		out.Members = append(out.Members, NewDevice(&elem))
	}

	// Default sort by ID
	sort.Slice(out.Members, func(i, j int) bool {
		return out.Members[i].ID < out.Members[j].ID
	})

	return // out
}

func ValidateDevice(r *DeviceResource) (m *model.Device, err error) {
	if r.Namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if r.DeviceID == "" {
		return nil, fmt.Errorf("deviceId is required")
	}
	if r.DeviceURI == "" {
		return nil, fmt.Errorf("deviceUri is required")
	}

	m = &model.Device{
		Namespace:      r.Namespace,
		DeviceID:       r.DeviceID,
		DeviceURI:      r.DeviceURI,
		SessionTimeout: r.SessionTimeout,
		PingInterval:   r.PingInterval,
		PongTimeout:    r.PongTimeout,
		EventsTopic:    r.EventsTopic,
	}

	return m, nil
}
