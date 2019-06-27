package connection

import (
	"fmt"
	"math/rand"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Controller struct {
	connections map[int32]*ControlChannel
	sessions    map[string]*Session
	sync.RWMutex
}

type Session struct {
	duplicateAllowed bool
	sessionIDs       []int32
}

type RegistrationDetails struct {
	SessionTimeout int    `json:"session_timeout,omitempty"`
	PingInterval   int    `json:"ping_interval,omitempty"`
	PongTimeout    int    `json:"pong_max_wait_time,omitempty"`
	EventsTopic    string `json:"events_topic,omitempty"`
}

func NewController() *Controller {
	return &Controller{
		connections: make(map[int32]*ControlChannel, 0),
		sessions:    make(map[string]*Session, 0),
	}
}

// RegisterControlChannel checks for existence of realm and returns on success
// a session ID and additional detials for the client.
func (ctrl *Controller) RegisterControlChannel(cc *ControlChannel, realm string) (int32, *RegistrationDetails, error) {
	if realm != "test" {
		return 0, nil, fmt.Errorf("ERR_NO_SUCH_REALM")
	}

	// Create a new session id and add control channel to active connections
	ctrl.Lock()
	defer ctrl.Unlock()

	sessID := random(1, 2^31)

	ctrl.connections[sessID] = cc
	ctrl.sessions[realm] = newSession(false, sessID)

	// TODO(DGL) Take the existing completeRegistration method of cc, add a
	// sync.RWMutex to ControlChannel struct and change the values inside the
	// method instead here!
	cc.sessionTimeout = 120

	log.Infof("controller add successfully a new control channel session with ID: %d", sessID)

	details := &RegistrationDetails{
		SessionTimeout: 120,
		PingInterval:   104,
		PongTimeout:    16,
		EventsTopic:    "iotcore.devicecontrol.events",
	}

	return sessID, details, nil
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
