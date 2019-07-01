package controlchannel

import (
	"net"
	"sync"
	"time"

	"github.com/nsyszr/lcm/pkg/devicecontrol/proto"
	log "github.com/sirupsen/logrus"
)

type Status int

const (
	StatusEstablished Status = iota
	StatusRegistered
)

type ControlChannel struct {
	sync.RWMutex
	mgr            *Manager
	conn           net.Conn
	status         Status
	lastMessageAt  time.Time
	stopCh         chan bool
	registeredCh   chan bool
	pingCh         chan bool
	realm          string
	sessionID      int32
	sessionTimeout int
	wsTerminateCh  chan<- struct{}
	wsCloseCh      chan struct{}
	wsOutboxCh     chan *Response
}

type Flag int

const (
	FlagContinue Flag = iota
	FlagCloseGracefully
	FlagTerminate
)

type Response struct {
	Flag Flag
	Data []byte
}

// New creates a control channel handler
func New(mgr *Manager, conn net.Conn, terminateCh chan<- struct{}) *ControlChannel {
	cc := &ControlChannel{
		mgr:           mgr,
		conn:          conn,
		status:        StatusEstablished,
		stopCh:        make(chan bool),
		registeredCh:  make(chan bool),
		pingCh:        make(chan bool),
		wsTerminateCh: terminateCh,
		wsCloseCh:     make(chan struct{}),
		wsOutboxCh:    make(chan *Response, 100),
	}

	go cc.inboxWorker()
	go cc.outboxWorker()

	// Start the go routine which ensures that registration happens within
	// given period.
	go cc.waitForReqistrationOrClose()

	return cc
}

// Close is called when the websocket handler method is exiting, e.g. the
// connection is closed.
func (cc *ControlChannel) Close() {
	// Tell our go waitForPingOrClose routines to stop listening for a signal
	cc.stopCh <- true
	// Unregister the control channel from the controller
	cc.mgr.Unregister(cc.sessionID)
}

// HandleMessage is called by the websocket handler when data is received from
// the connected client.
func (cc *ControlChannel) HandleMessage(data []byte) ([]byte, Flag, error) {
	log.Infof("controlchannel handles message '%s'", string(data))

	// Unmarshal the message to get the message type for further processing.
	msgType, msg, err := proto.UnmarshalMessage(data)
	if err != nil {
		return cc.terminateAndLogError("invalid payload", err)
	}

	switch msgType {
	case proto.MessageTypeHello:
		return cc.handleMessage(msg, cc.helloHandler())
	case proto.MessageTypePing:
		return cc.handleMessage(msg, cc.ensureRegistered(cc.keepAliveHandler()))
	case proto.MessageTypePublish:
		return cc.handleMessage(msg, cc.ensureRegistered(cc.eventHandler()))
	}

	return cc.terminateAndLog("unhandled message")
}

// AdmitRegistration is called by the controller after successful registration
// (authorization) of the client. This method sets neccessary values for
// running the control channel and starts the keep alive handling in the
// background (waitForPingOrClose).
func (cc *ControlChannel) AdmitRegistration(sessionID int32, realm string, sessionTimeout int) {
	// The current state is changing! Lock the access to the control channel
	// object until we're finished.
	// cc.Lock()
	// defer cc.Unlock()

	cc.Lock()
	cc.status = StatusRegistered
	cc.sessionID = sessionID
	cc.realm = realm
	cc.sessionTimeout = sessionTimeout
	cc.Unlock()

	// Start the session timeout timer. If client doesn't send a ping withing
	// given timeout the connection will be closed.
	go cc.waitForPingOrClose()

	log.Infof("controlchannel registered successful for device '%s'", realm)
}

func (cc *ControlChannel) waitForReqistrationOrClose() {
	log.Info("controlchannel waitForReqistrationOrClose method started")
	for {
		select {
		case <-cc.registeredCh:
			log.Info("controlchannel waitForReqistrationOrClose method successfully received registration signal")
			return
		case <-cc.stopCh:
			log.Info("controlchannel waitForReqistrationOrClose method received stop signal")
			return
		case <-time.After(10 * time.Second): // TODO: get timeout from config
			log.Warn("controlchannel waitForReqistrationOrClose method timed out and terminates the connection")
			// Close the session, since it's not registered within time
			close(cc.wsCloseCh)
			return
		}
	}
}

// messageHandler is a tooling for handling incoming messages. It is similar
// to the go http handler implementation. It allows us to create middleware
// handlers, e.g. the ensureRegistered handler.
type messageHandler interface {
	Handle(msg interface{}) ([]byte, Flag, error)
}

type messageHandlerFunc func(msg interface{}) ([]byte, Flag, error)

func (f messageHandlerFunc) Handle(msg interface{}) ([]byte, Flag, error) {
	return f(msg)
}

// handleMessage is the main method that is called by the public HandleMessage
// function. It expects a handler of interface messageHandler. This method is
// similar to the go implementation of http.HandleFunc.
func (cc *ControlChannel) handleMessage(msg interface{}, h messageHandler) ([]byte, Flag, error) {
	// We lock the access to control channel object until we handled the
	// complete message. This ensures that we can safely modify the object and
	// that the current state isn't touched meanwhile.
	cc.Lock()
	cc.lastMessageAt = time.Now()
	cc.Unlock()

	return h.Handle(msg)
}

func (cc *ControlChannel) helloHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		helloMsg, err := proto.MustHelloMessage(msg)
		if err != nil {
			return cc.terminateAndLogError("hello message expected", err)
		}

		// Notify the waitForReqistrationOrClose go routine that we're about to
		// register the connection, otherwise the connection can be closed
		// during registration.
		cc.registeredCh <- true

		sessID, details, err := cc.mgr.Register(cc, helloMsg.Realm)
		if err != nil && IsRegistrationError(err) {
			log.Warnf("controlchannel rejected for device '%s'", helloMsg.Realm)
			e := err.(*RegistrationError)
			return cc.abortMessageAndClose(e.Reason, e.Details)
		} else if err != nil {
			log.Errorf("controlchannel registration failed: %s", err.Error())
			return cc.terminateAndLogError("could not register controlchannel", err)
		}

		return cc.welcomeMessage(sessID, details)
	})
}

func (cc *ControlChannel) waitForPingOrClose() {
	log.Info("controlchannel waitForPingOrClose method started")
	for {
		select {
		case <-cc.pingCh:
			log.Info("controlchannel waitForPingOrClose method successfully received ping signal")
			// We do not exit the loop because we reset the timeout only
		case <-cc.stopCh:
			log.Info("controlchannel waitForPingOrClose method received stop signal")
			return
		case <-time.After(time.Duration(cc.sessionTimeout) * time.Second):
			log.Warn("controlchannel waitForPingOrClose method timed out and terminates the connection")
			// Close the session, since it doesn't reponds within given period
			close(cc.wsCloseCh)
			return
		}
	}
}

func (cc *ControlChannel) ensureRegistered(next messageHandler) messageHandler {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		if cc.status != StatusRegistered {
			return cc.terminateAndLog("controlchannel is not registered")
		}
		return next.Handle(msg)
	})
}

func (cc *ControlChannel) keepAliveHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		// Notify the waitForPingOrClose method that we received a ping,
		// otherwise session timeout occurs and closes the connection.
		go func() {
			cc.pingCh <- true
		}()

		return cc.pongMessage()
	})
}

func (cc *ControlChannel) eventHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		publishMsg, err := proto.MustPublishMessage(msg)
		if err != nil {
			return cc.terminateAndLogError("publish message expected", err)
		}

		return cc.publishedMessage(publishMsg.RequestID, 0)
	})
}

func (cc *ControlChannel) terminateAndLog(message string) ([]byte, Flag, error) {
	log.Errorf("controlchannel terminates with message: %s", message)
	cc.pushBackMessage(FlagTerminate, nil)
	return nil, FlagTerminate, nil
}

func (cc *ControlChannel) terminateAndLogError(message string, err error) ([]byte, Flag, error) {
	log.Errorf("controlchannel terminates with message and error: %s: %s", message, err.Error())
	cc.pushBackMessage(FlagTerminate, nil)
	return nil, FlagTerminate, nil
}

func (cc *ControlChannel) abortMessageAndClose(reason string, details interface{}) ([]byte, Flag, error) {
	out, err := proto.MarshalNewAbortMessage(reason, details)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		return cc.terminateAndLogError("could not marshal message", err)
	}
	cc.pushBackMessage(FlagCloseGracefully, out)
	return out, FlagCloseGracefully, nil
}

func (cc *ControlChannel) welcomeMessage(sessionID int32, details interface{}) ([]byte, Flag, error) {
	out, err := proto.MarshalNewWelcomeMessage(sessionID, details)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		return cc.terminateAndLogError("could not marshal message", err)
	}
	cc.pushBackMessage(FlagContinue, out)
	return out, FlagContinue, nil
}

func (cc *ControlChannel) pongMessage() ([]byte, Flag, error) {
	out, err := proto.MarshalNewPongMessage()
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		return cc.terminateAndLogError("could not marshal message", err)
	}
	cc.pushBackMessage(FlagContinue, out)
	return out, FlagContinue, nil
}

func (cc *ControlChannel) publishedMessage(requestID, publicationID int32) ([]byte, Flag, error) {
	out, err := proto.MarshalNewPublishedMessage(requestID, publicationID)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		return cc.terminateAndLogError("could not marshal message", err)
	}
	cc.pushBackMessage(FlagContinue, out)
	return out, FlagContinue, nil
}

func (cc *ControlChannel) pushBackMessage(flag Flag, data []byte) bool {
	select {
	case cc.wsOutboxCh <- newResponse(flag, data):
		return true
	default:
		return false // Buffer is full
	}
}

func newResponse(flag Flag, data []byte) *Response {
	r := &Response{
		Flag: flag,
	}
	if data != nil {
		r.Data = make([]byte, len(data))
		copy(r.Data, data)
	}
	return r
}
