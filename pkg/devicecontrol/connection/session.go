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
		log.Warnf("session terminates immediately because of invalid payload: %s", err.Error())
		return nil, FlagTerminate, nil
	}

	// Update LastMessageAt
	sess.LastMessageAt = time.Now()

	switch msgType {
	case proto.MessageTypeHello:
		return sess.handleMessage(msg, sess.helloHandler())
	case proto.MessageTypePing:
		return sess.handleMessage(msg, sess.ensureRegistered(sess.keepAliveHandler()))
	}

	log.Error("session terminates immediately because of unhandled message")
	return nil, FlagTerminate, nil
}

func (sess *Session) handleMessage(msg interface{}, h messageHandler) ([]byte, Flag, error) {
	return h.Handle(msg)
}

func (sess *Session) helloHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		helloMsg, err := proto.MustHelloMessage(msg)
		if err != nil {
			log.Infof("session terminates immediately because of protocol violation: %s", err.Error())
			return nil, FlagTerminate, nil
		}

		// Do authentication
		if helloMsg.Realm != "test" {
			out, err := proto.MarshalNewAbortMessage("ERR_NO_SUCH_REALM", nil)
			// This error should happen never! If it happens log an urgent error
			// and terminate the websocket session for safety.
			if err != nil {
				log.Errorf("could not marshal a message: %s", err.Error())
				return nil, FlagTerminate, nil
			}

			return out, FlagCloseGracefully, nil
		}

		// Send welcome message
		out, err := proto.MarshalNewWelcomeMessage(1234, nil)
		// This error should happen never! If it happens log an urgent error
		// and terminate the websocket session for safety.
		if err != nil {
			log.Errorf("could not marshal a message: %s", err.Error())
			return nil, FlagTerminate, nil
		}

		sess.Status = SessionStatusRegistered
		log.Info("session responds with WELCOME")

		return out, FlagContinue, nil
	})
}

func (sess *Session) ensureRegistered(next messageHandler) messageHandler {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		if sess.Status != SessionStatusRegistered {
			return nil, FlagTerminate, nil
		}
		return next.Handle(msg)
	})
}

func (sess *Session) keepAliveHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) ([]byte, Flag, error) {
		out, err := proto.MarshalNewPongMessage()
		if err != nil {
			// TODO(DGL) Convert error to something useful
			log.Errorf("session terminates immediately because of marshalling failed: %s", err.Error())
			return nil, FlagTerminate, nil
		}

		return out, FlagContinue, nil
	})
}
