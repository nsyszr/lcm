package resource

import (
	"sort"
	"time"

	"github.com/nsyszr/lcm/pkg/model"
)

type SessionResource struct {
	ID             int32     `json:"id"`
	Namespace      string    `json:"namespace"`
	DeviceID       string    `json:"deviceId"`
	DeviceURI      string    `json:"deviceUri"`
	SessionTimeout int       `json:"sessionTimeout"`
	LastMessageAt  time.Time `json:"lastMessageAt"`
}

type SessionListResource struct {
	Members []*SessionResource `json:"members"`
}

func NewSession(m *model.Session) (out *SessionResource) {
	out = &SessionResource{
		ID:             m.ID,
		Namespace:      m.Namespace,
		DeviceID:       m.DeviceID,
		DeviceURI:      m.DeviceURI,
		SessionTimeout: m.SessionTimeout,
		LastMessageAt:  m.LastMessageAt,
	}

	return // out
}

func NewSessionList(m map[int32]model.Session) (out *SessionListResource) {
	out = &SessionListResource{
		Members: make([]*SessionResource, 0),
	}

	for _, elem := range m {
		out.Members = append(out.Members, NewSession(&elem))
	}

	// Default sort by ID
	sort.Slice(out.Members, func(i, j int) bool {
		return out.Members[i].ID < out.Members[j].ID
	})

	return // out
}
