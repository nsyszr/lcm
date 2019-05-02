package devicecontrol

import (
	"strings"

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
func (sess *session) handle(req []byte) ([]byte, bool, error) {
	log.Infof("We've got a message: %s", string(req))

	msg := string(req)

	switch sess.state {
	case StateEstablished:
		if !strings.HasPrefix(msg, "HELLO:") {
			// Quit the connection immediately because we received shit!
			return nil, true, nil
		}

		realm := msg[6:]
		if realm != "test" {
			return []byte("ABORT:ERR_NO_SUCH_REALM"), true, nil
		}

		sess.state = StateRegistered
		return []byte("WELCOME"), false, nil
	case StateRegistered:
		return []byte("ECHO:" + msg), false, nil
	}

	// This shouldn't be the case, it's better to close the connection
	return nil, true, nil
}
