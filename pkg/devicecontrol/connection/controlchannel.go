package connection

import (
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/nsyszr/lcm/pkg/devicecontrol/proto"
	log "github.com/sirupsen/logrus"
)

type ConnectionStatus int

const (
	ConnectionStatusEstablished ConnectionStatus = iota
	ConnectionStatusRegistered
)

type ControlChannel struct {
	sync.RWMutex
	ctrl           *Controller
	conn           net.Conn
	w              *wsutil.Writer
	lastMessageAt  time.Time
	status         ConnectionStatus
	stopCh         chan bool
	registeredCh   chan bool
	pingCh         chan bool
	realm          string
	sessionTimeout int
}

type Flag int

const (
	FlagContinue Flag = iota
	FlagCloseGracefully
	FlagTerminate
)

func NewControlChannel(ctrl *Controller, conn net.Conn, w *wsutil.Writer) *ControlChannel {
	cc := &ControlChannel{
		ctrl:         ctrl,
		conn:         conn,
		w:            w,
		status:       ConnectionStatusEstablished,
		stopCh:       make(chan bool),
		registeredCh: make(chan bool),
		pingCh:       make(chan bool),
	}

	// Start the go routine which ensures that registration happens within
	// given period.
	go cc.waitForReqistrationOrClose()

	return cc
}

func (cc *ControlChannel) Close() {
	// Tell our go routines to stop listening for a signal
	cc.stopCh <- true
}

func (cc *ControlChannel) HandleMessage(data []byte) ([]byte, Flag, error) {
	log.Infof("controlchannel handles message '%s'", string(data))

	// Unmarshal the message to get the message type for further processing.
	msgType, msg, err := proto.UnmarshalMessage(data)
	if err != nil {
		return terminateAndLogError("invalid payload", err)
	}

	switch msgType {
	case proto.MessageTypeHello:
		return cc.handleMessage(msg, cc.helloHandler())
	case proto.MessageTypePing:
		return cc.handleMessage(msg, cc.ensureRegistered(cc.keepAliveHandler()))
	case proto.MessageTypePublish:
		return cc.handleMessage(msg, cc.ensureRegistered(cc.eventHandler()))
	}

	return terminateAndLog("unhandled message")
}

func (cc *ControlChannel) AdmitRegistration(realm string, sessionTimeout int) {
	// The current state is changing! Lock the access to the control channel
	// object until we're finished.
	// cc.Lock()
	// defer cc.Unlock()

	cc.Lock()
	cc.status = ConnectionStatusRegistered
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
			cc.closeWebSocket()
			return
		}
	}
}

func (cc *ControlChannel) closeWebSocket() {
	cc.w.Reset(cc.conn, ws.StateServerSide, ws.OpClose)

	// Write empty string
	var err error
	if _, err = cc.w.Write([]byte("")); err == nil {
		err = cc.w.Flush()
	}
	if err != nil {
		// TODO We should attach this information to the device log perhaps.
		log.Errorf("controlchannel websocket write error: %s", err)
	}
}

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
			return terminateAndLogError("hello message expected", err)
		}

		// Notify the waitForReqistrationOrClose go routine that we're about to
		// register the connection, otherwise the connection will be closed.
		cc.registeredCh <- true

		sessID, details, err := cc.ctrl.RegisterControlChannel(cc, helloMsg.Realm)
		if err != nil {
			log.Warnf("controlchannel rejected for device '%s'", helloMsg.Realm)
			return abortMessageAndClose("ERR_NO_SUCH_REALM", nil)
		}

		return welcomeMessage(sessID, details)
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
			cc.closeWebSocket()
			return
		}
	}
}

func (cc *ControlChannel) ensureRegistered(next messageHandler) messageHandler {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		if cc.status != ConnectionStatusRegistered {
			return terminateAndLog("controlchannel is not registered")
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

		return pongMessage()
	})
}

func (cc *ControlChannel) eventHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		publishMsg, err := proto.MustPublishMessage(msg)
		if err != nil {
			return terminateAndLogError("publish message expected", err)
		}

		return publishedMessage(publishMsg.RequestID, 0)
	})
}

func terminateAndLog(message string) ([]byte, Flag, error) {
	log.Errorf("controlchannel terminates with message: %s", message)
	return nil, FlagTerminate, nil
}

func terminateAndLogError(message string, err error) ([]byte, Flag, error) {
	log.Errorf("controlchannel terminates with message and error: %s: %s", message, err.Error())
	return nil, FlagTerminate, nil
}

func abortMessageAndClose(reason string, details interface{}) ([]byte, Flag, error) {
	out, err := proto.MarshalNewAbortMessage(reason, details)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		return terminateAndLogError("could not marshal message", err)
	}
	return out, FlagCloseGracefully, nil
}

func welcomeMessage(sessionID int32, details interface{}) ([]byte, Flag, error) {
	out, err := proto.MarshalNewWelcomeMessage(sessionID, details)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		return terminateAndLogError("could not marshal message", err)
	}
	return out, FlagContinue, nil
}

func pongMessage() ([]byte, Flag, error) {
	out, err := proto.MarshalNewPongMessage()
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		return terminateAndLogError("could not marshal message", err)
	}
	return out, FlagContinue, nil
}

func publishedMessage(requestID, publicationID int32) ([]byte, Flag, error) {
	out, err := proto.MarshalNewPublishedMessage(requestID, publicationID)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		return terminateAndLogError("could not marshal message", err)
	}
	return out, FlagContinue, nil
}
