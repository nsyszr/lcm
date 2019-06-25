package connection

import (
	"time"

	"github.com/gobwas/ws/wsutil"
	"github.com/nsyszr/lcm/pkg/devicecontrol/proto"
	log "github.com/sirupsen/logrus"
)

type SessionStatus int

const (
	SessionStatusEstablished SessionStatus = iota
	SessionStatusRegistered
)

type Session struct {
	Controller    *Controller
	Writer        *wsutil.Writer
	LastMessageAt time.Time
	Status        SessionStatus
}

type Flag int

const (
	FlagContinue Flag = iota
	FlagCloseGracefully
	FlagTerminate
)

func newSession(ctrl *Controller, w *wsutil.Writer) *Session {
	return &Session{
		Controller: ctrl,
		Writer:     w,
		Status:     SessionStatusEstablished,
	}
}

func (sess *Session) HandleMessage(data []byte) ([]byte, Flag, error) {
	log.Infof("session handle message '%s'", string(data))

	// Unmarshal the message to get the message type for further processing.
	msgType, msg, err := proto.UnmarshalMessage(data)
	if err != nil {
		return terminateAndLogError("invalid payload", err)
	}

	switch msgType {
	case proto.MessageTypeHello:
		return sess.handleMessage(msg, sess.helloHandler())
	case proto.MessageTypePing:
		return sess.handleMessage(msg, sess.ensureRegistered(sess.keepAliveHandler()))
	}

	return terminateAndLog("unhandled message")
}

func (sess *Session) handleMessage(msg interface{}, h messageHandler) ([]byte, Flag, error) {
	sess.LastMessageAt = time.Now()
	return h.Handle(msg)
}

func (sess *Session) helloHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		helloMsg, err := proto.MustHelloMessage(msg)
		if err != nil {
			return terminateAndLogError("hello message expected", err)
		}

		// Do authentication
		if helloMsg.Realm != "test" {
			return abortMessageAndClose("ERR_NO_SUCH_REALM", nil)
		}

		sess.Status = SessionStatusRegistered
		log.Info("session responds with WELCOME")

		return welcomeMessage(1234, nil)
	})
}

func (sess *Session) ensureRegistered(next messageHandler) messageHandler {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		if sess.Status != SessionStatusRegistered {
			return terminateAndLog("session is not registered")
		}
		return next.Handle(msg)
	})
}

func (sess *Session) keepAliveHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		return pongMessage()
	})
}

func terminateAndLog(message string) ([]byte, Flag, error) {
	log.Errorf("devicecontrol: %s", message)
	return nil, FlagTerminate, nil
}

func terminateAndLogError(message string, err error) ([]byte, Flag, error) {
	log.Errorf("devicecontrol: %s: %s", message, err.Error())
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
	out, err := proto.MarshalNewWelcomeMessage(1234, nil)
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
