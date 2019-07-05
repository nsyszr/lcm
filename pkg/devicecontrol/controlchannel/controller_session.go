package controlchannel

import (
	"strings"
	"time"

	"github.com/nsyszr/lcm/pkg/model"
	log "github.com/sirupsen/logrus"
)

// RegisterSession checks first for existence of realm and on success it's starts a
// new session, returns the session ID and details that are sent to the client.
func (ctrl *Controller) RegisterSession(cc *ControlChannel, realm string) (int32, interface{}, error) {
	if realm != "test@test" {
		return 0, nil, NewRegistrationError(ErrReasonNoSuchRelam, nil)
	}

	deviceIDAndURI := strings.SplitN(realm, "@", 2)
	log.Infof("deviceIDAndURI=%v", deviceIDAndURI)
	if len(deviceIDAndURI) != 2 {
		return 0, nil, NewRegistrationError(ErrReasonNoSuchRelam, nil)
	}

	// Check if session exists
	// TODO(DGL) Fix hardcoded namespace
	_, err := ctrl.store.Sessions().FindByNamespaceAndDeviceID("default", deviceIDAndURI[0])
	if err != nil && err.Error() != "not found" {
		return 0, nil, NewTechnicalExceptionError(nil)
	}
	if err == nil {
		log.Warnf("controller rejected the control channel becuase session for '%s' exists already", deviceIDAndURI[0])
		return 0, nil, NewRegistrationError("ERR_SESSION_EXISTS", nil)
	}

	// Create a new session in the store
	// TODO(DGL) Fix hardcoded namespace
	sess := model.Session{
		Namespace:     "default",
		DeviceID:      deviceIDAndURI[0],
		DeviceURI:     deviceIDAndURI[1],
		LastMessageAt: time.Now().Round(time.Second).UTC(),
		Timeout:       120,
	}
	if err := ctrl.store.Sessions().Create(&sess); err != nil {
		return 0, nil, NewTechnicalExceptionError(nil)
	}

	// Add session to controller
	// ctrl.Lock()
	// ctrl.connections[sess.ID] = cc
	// ctrl.sessions[realm] = newSession(false, m.ID)
	// ctrl.Unlock()

	log.Infof("controller added successfully a new control channel session with ID: %d", sess.ID)

	// Tell control channel that the registration is admitted
	cc.AdmitRegistration(sess.ID, 120, realm)

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
		EventsTopic:    "device",
	}
	return sess.ID, details, nil
}

// UnregisterSession removes a session from the connection and session list.
func (ctrl *Controller) UnregisterSession(sessionID int32) {
	// Do something
	ctrl.store.Sessions().Delete(sessionID)
	log.Infof("controller removed successfully the control channel session with ID: %d", sessionID)
}
