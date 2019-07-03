package controlchannel

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/storage"
	log "github.com/sirupsen/logrus"
)

type Controller struct {
	nc             *nats.Conn
	store          storage.Interface
	messageTimeout int
}

func NewController(nc *nats.Conn, store storage.Interface) *Controller {
	return &Controller{
		nc:             nc,
		store:          store,
		messageTimeout: 16,
	}
}

func (ctrl *Controller) Subscribe() error {
	if ctrl.nc == nil {
		return fmt.Errorf("controller: connection to nats is missing")
	}

	if _, err := ctrl.nc.QueueSubscribe("iotcore.devicecontrol.v1.*.publish", "iotcore.devicecontrol.v1.queue.publish", func(msg *nats.Msg) {
		if err := ctrl.handlePublishRequest(msg); err != nil {
			log.Error("controller failed to handle publish request: ", err.Error())
		}
	}); err != nil {
		return err
	}

	if _, err := ctrl.nc.QueueSubscribe("iotcore.devicecontrol.v1.*.call", "iotcore.devicecontrol.v1.queue.call", func(msg *nats.Msg) {
		if err := ctrl.handleCallRequest(msg); err != nil {
			log.Error("controller failed to handle call request: ", err.Error())
		}
	}); err != nil {
		return err
	}

	return nil
}

// NewControlChannel creates a control channel handler
func (ctrl *Controller) NewControlChannel(conn net.Conn, terminateCh chan<- struct{}) *ControlChannel {
	cc := &ControlChannel{
		ctrl:           ctrl,
		nc:             ctrl.nc,
		conn:           conn,
		status:         StatusEstablished,
		stopCh:         make(chan bool),
		registeredCh:   make(chan bool),
		pingCh:         make(chan bool),
		wsTerminateCh:  terminateCh,
		wsCloseCh:      make(chan struct{}),
		wsOutboxCh:     make(chan *Response, 100),
		nextRequestID:  1,
		resultChannels: make(map[int32]chan<- interface{}),
	}

	go cc.inboxWorker()
	go cc.outboxWorker()

	// Start the go routine which ensures that registration happens within
	// given period.
	go cc.waitForReqistrationOrClose()

	return cc
}

func (ctrl *Controller) replyMessage(replyTo string, rep interface{}) error {
	data, err := json.Marshal(rep)
	if err != nil {
		return err
	}

	return ctrl.nc.Publish(replyTo, data)
}
