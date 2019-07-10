package model

import "time"

// Device is a model of the persistency layer
type Device struct {
	ID             int32
	Namespace      string
	DeviceID       string
	DeviceURI      string
	SessionTimeout int
	PingInterval   int
	PongTimeout    int
	EventsTopic    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
