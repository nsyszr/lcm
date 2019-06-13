package devicecontrol

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsyszr/lcm/pkg/devicecontrol/proto"
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
	id    int32
	state State
}

func init() {
	log.Debug("devicecontrol/session loaded")
	// Initialize the random seeder
	rand.Seed(time.Now().UnixNano())
}

func newSession(h *Handler) (*session, error) {
	sess := &session{
		h:     h,
		state: StateEstablished,
	}

	// Find a unique session ID within a period of 10 seconds and append the
	// session to the handler. Otherwise return an error.
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			log.Error("could not find a unique session id for new session")
			return nil, fmt.Errorf("devicecontrol: could not find a unique session id")
		default:
			id := random(1, 2^31)
			log.Debugf("propose session ID: %d", id)
			h.Lock()
			if _, ok := h.sessions[id]; !ok {
				sess.id = id
				h.sessions[id] = sess
				h.Unlock()
				return sess, nil
			}
			h.Unlock()
		}
	}
}

func random(min int32, max int32) int32 {
	return int32(rand.Int31n(max-min) + min)
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

// handle method processes the received message from the websocket connection.
// It returns valid payload to send back to the client or nil. If the
// connection should be closed it returns true as second return value. In
// case of an error the third return value is not nil. The connection has to
// be terminated, too.
func (sess *session) handle(req []byte) ([]byte, Flag, error) {
	log.Infof("session handles message: %s", string(req))

	// Unmarshal the message to get the message type for further processing.
	msgType, msg, err := proto.UnmarshalMessage(req)
	if err != nil {
		log.Infof("session quits immediately because of invalid payload: %s", err.Error())
		return nil, FlagTerminate, nil
	}

	// Depending on the state we handle the messages. In case of an established
	// session we expect a hello message to authorize the client. Otherwise we
	// are processing the message against the interested messaging parties,
	// except the ping. The ping / pong keepalive is handled by the session.
	switch sess.state {
	case StateEstablished:
		if msgType != proto.MessageTypeHello {
			// Quit the connection immediately because of protocol violation
			log.Info("session quits immediately because of protocol violation: expected a hello message")
			return nil, FlagTerminate, nil
		}
		helloMsg, ok := msg.(proto.HelloMessage)
		// This error should happen never! If it happens log an urgent error
		// and terminate the websocket session for safety.
		if !ok {
			log.Error("type cast the unmarshalled hello message to a hello message type failed")
			return nil, FlagTerminate, nil
		}

		if helloMsg.Realm != "test" {
			log.Info("session quits with ABORT:ERR_NO_SUCH_REALM")

			abortMsg := proto.AbortMessage{Reason: "ERR_NO_SUCH_REALM", Details: nil}
			res, err := proto.MarshalMessage(abortMsg)

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

		welcomeMsg := proto.WelcomeMessage{SessionID: sess.id, Details: nil}
		res, err := proto.MarshalMessage(welcomeMsg)

		// This error should happen never! If it happens log an urgent error
		// and terminate the websocket session for safety.
		if err != nil {
			log.Errorf("could not marshal a message: %s", err.Error())
			return nil, FlagTerminate, nil
		}

		return res, FlagContinue, nil
	case StateRegistered:
		if msgType == proto.MessageTypePing {
			log.Info("session responds with PING")

			pongMsg := proto.PongMessage{Details: nil}
			res, err := proto.MarshalMessage(pongMsg)

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
