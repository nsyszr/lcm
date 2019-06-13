package devicecontrol

import (
	"github.com/nsyszr/lcm/pkg/devicecontrol/message"
	log "github.com/sirupsen/logrus"
)

type State int

const (
	StateEstablished State = iota
	StateAnnounced
	StateRegistered
	StateAborted
	StateClosed
)

type Flag int

const (
	FlagContinue Flag = iota
	FlagCloseGracefully
	FlagTerminate
)

func (state State) String() string {
	names := []string{
		"ESTABLISHED",
		"ANNOUNCED",
		"REGISTERED",
		"ABORTED",
		"CLOSED"}

	if state < StateEstablished || state > StateClosed {
		return "UNKNOWN"
	}

	return names[state]
}

type session struct {
	h     *Handler
	id    int
	state State
}

func newSession(h *Handler) *session {
	return &session{
		h:     h,
		state: StateEstablished,
	}
}

func (sess *session) close() {
	// Remove the session from sessions table if the session has a session id
	if sess.id != 0 {

		// TODO add a proper state handling here, to clean up the session in
		// the database, to deregister it from the app, etc.

		sess.h.Lock()
		defer sess.h.Unlock()
		delete(sess.h.sessions, sess.id)
	}
}

// handle processes the received message from the websocket connection.
// It returns valid payload to send back to the client or nil. If the
// connection should be closed it returns true as second return value. In
// case of an error the third return value is not nil. The connection has to
// be terminated, too.
func (sess *session) handle(req []byte) ([]byte, Flag, error) {
	log.Infof("session handles message: %s", string(req))

	msgType, msg, err := message.Unmarshal(req)
	if err != nil {
		log.Infof("session quits immediately because of invalid payload: %s", err.Error())
		return nil, FlagTerminate, nil
	}

	switch sess.state {
	case StateEstablished:
		if msgType != message.MessageTypeHello {
			// Quit the connection immediately because of protocol violation
			log.Info("session quits immediately because of protocol violation: expected a hello message")
			return nil, FlagTerminate, nil
		}
		helloMsg, _ := msg.(message.HelloMessage)

		if helloMsg.Realm != "test" {
			log.Info("session quits with ABORT:ERR_NO_SUCH_REALM")

			abortMsg := message.AbortMessage{Reason: "ERR_NO_SUCH_REALM", Details: nil}
			res, err := message.Marshal(abortMsg)

			// This error should happen never! If it happens log an urgent error
			// and terminate the websocket session for safety.
			if err != nil {
				log.Errorf("could not marshal a message: %s", err.Error())
				return nil, FlagTerminate, nil
			}

			return res, FlagCloseGracefully, nil
		}

		sess.state = StateRegistered
		log.Info("session responds with WELCOME")

		welcomeMsg := message.WelcomeMessage{SessionID: 1234, Details: nil}
		res, err := message.Marshal(welcomeMsg)

		// This error should happen never! If it happens log an urgent error
		// and terminate the websocket session for safety.
		if err != nil {
			log.Errorf("could not marshal a message: %s", err.Error())
			return nil, FlagTerminate, nil
		}

		return res, FlagContinue, nil
	case StateRegistered:
		if msgType == message.MessageTypePing {
			log.Info("session responds with PING")

			pongMsg := message.PongMessage{Details: nil}
			res, err := message.Marshal(pongMsg)

			// This error should happen never! If it happens log an urgent error
			// and terminate the websocket session for safety.
			if err != nil {
				log.Errorf("could not marshal a message: %s", err.Error())
				return nil, FlagTerminate, nil
			}

			return res, FlagContinue, nil
		}
	}

	// This shouldn't be the case, it's better to close the connection
	return nil, FlagTerminate, nil
}
