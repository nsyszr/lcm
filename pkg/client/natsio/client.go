package natsio

import (
	"fmt"

	nats "github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/client"
)

type natsClient struct {
	cfg *Config
	nc  *nats.Conn
}

func New(cfg *Config) (client.Interface, error) {
	nc, err := nats.Connect(cfg.url)
	if err != nil {
		return nil, err
	}
	return &natsClient{
		cfg: cfg,
		nc:  nc,
	}, nil
}

func (c *natsClient) RequestCommand(id, cmd string, args []byte) ([]byte, error) {
	subj := fmt.Sprintf("%s.target.%s.command.%s", c.cfg.baseSubject, id, cmd)
	msg, err := c.nc.Request(subj, args, c.cfg.defaultTimeout)
	if err != nil {
		return nil, err
	}

	return msg.Data, nil
}
