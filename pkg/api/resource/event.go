package resource

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/nsyszr/lcm/pkg/model"
)

/*
type Event struct {
	ID         int32
	Namespace  string
	SourceType string
	SourceID   string
	Topic      string
	Timestamp  time.Time
	Details    string

	CreatedAt time.Time
	UpdatedAt time.Time
}*/

type EventResource struct {
	ID         int32       `json:"id"`
	Namespace  string      `json:"namespace"`
	SourceType string      `json:"sourceType"`
	SourceID   string      `json:"sourceId"`
	Topic      string      `json:"topic"`
	Timestamp  time.Time   `json:"timestamp"`
	Details    interface{} `json:"details"`
}

type EventListResource struct {
	Members []*EventResource `json:"members"`
}

func NewEvent(m *model.Event) (out *EventResource) {
	out = &EventResource{
		ID:         m.ID,
		Namespace:  m.Namespace,
		SourceType: m.SourceType,
		SourceID:   m.SourceID,
		Topic:      m.Topic,
		Timestamp:  m.Timestamp,
	}

	var details interface{}
	if err := json.Unmarshal([]byte(m.Details), &details); err == nil {
		out.Details = details
	}

	return // out
}

func NewEventList(m map[int32]model.Event) (out *EventListResource) {
	out = &EventListResource{
		Members: make([]*EventResource, 0),
	}

	for _, elem := range m {
		out.Members = append(out.Members, NewEvent(&elem))
	}

	// Default sort by ID
	sort.Slice(out.Members, func(i, j int) bool {
		return out.Members[i].ID < out.Members[j].ID
	})

	return // out
}
