package controlchannel

import (
	"math/rand"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	connections map[int32]*ControlChannel
	sessions    map[string]*Session
	sync.RWMutex
}

type Session struct {
	duplicateAllowed bool
	sessionIDs       []int32
}

func NewManager() *Manager {
	return &Manager{
		connections: make(map[int32]*ControlChannel, 0),
		sessions:    make(map[string]*Session, 0),
	}
}

// Register checks first for existence of realm and on success it's starts a
// new session, returns the session ID and details that are sent to the client.
func (mgr *Manager) Register(cc *ControlChannel, realm string) (int32, interface{}, error) {
	if realm != "test@test" {
		return 0, nil, NewRegistrationError(ErrReasonNoSuchRelam, nil)
	}

	// Create a new session id and add control channel to active connections
	// TODO(DGL) Add loop for unique session ID
	sessID := random(1, 2^31)
	mgr.Lock()
	mgr.connections[sessID] = cc
	mgr.sessions[realm] = newSession(false, sessID)
	mgr.Unlock()
	log.Infof("controller add successfully a new control channel session with ID: %d", sessID)

	// Tell control channel that the registration is admitted
	cc.AdmitRegistration(sessID, realm, 120)

	// Return the results of the registration to the control channel
	type registrationDetails struct {
		SessionTimeout int    `json:"session_timeout,omitempty"`
		PingInterval   int    `json:"ping_interval,omitempty"`
		PongTimeout    int    `json:"pong_max_wait_time,omitempty"`
		EventsTopic    string `json:"events_topic,omitempty"`
	}

	details := &registrationDetails{
		SessionTimeout: 20,
		PingInterval:   16,
		PongTimeout:    4,
		EventsTopic:    "iotcore.devicecontrol.events",
	}
	return sessID, details, nil
}

// Unregister removes a session from the connection and session list.
func (mgr *Manager) Unregister(sessionID int32) {
	// Do something
}

func newSession(duplicateAllowed bool, sessionID int32) *Session {
	sess := &Session{
		duplicateAllowed: duplicateAllowed,
		sessionIDs:       make([]int32, 1),
	}
	sess.sessionIDs = append(sess.sessionIDs, sessionID)
	return sess
}

func random(min int32, max int32) int32 {
	return int32(rand.Int31n(max-min) + min)
}
