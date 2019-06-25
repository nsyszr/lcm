package natsio

import "time"

type Config struct {
	url            string
	baseSubject    string
	defaultTimeout time.Duration
}
