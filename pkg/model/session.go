package model

import "time"

// Session is a model of the persistency layer
type Session struct {
	ID            int32
	Namespace     string
	DeviceID      string
	DeviceURI     string
	LastMessageAt time.Time
	Timeout       int

	CreatedAt time.Time
	UpdatedAt time.Time
}
