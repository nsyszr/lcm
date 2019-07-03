package model

import "time"

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
}
