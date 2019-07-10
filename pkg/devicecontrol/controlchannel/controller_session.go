package controlchannel

import (
	"fmt"
	"strings"
	"time"

	"github.com/nsyszr/lcm/pkg/devicecontrol/proto"
	"github.com/nsyszr/lcm/pkg/model"
	"github.com/nsyszr/lcm/pkg/storage"
	log "github.com/sirupsen/logrus"
)

// RegisterSession checks first for existence of realm and on success it's starts a
// new session, returns the session ID and details that are sent to the client.
func (ctrl *Controller) RegisterSession(cc *ControlChannel, realm string) (int32, interface{}, error) {
	deviceIDAndURI := strings.SplitN(realm, "@", 2)
	if len(deviceIDAndURI) != 2 {
		return 0, nil, proto.NewRegistrationError(proto.ErrReasonNoSuchRelam,
			fmt.Sprintf("realm '%s' is not valid", realm))
	}

	// Find the device
	// TODO(DGL) Fix hardcoded namespace
	device, err := ctrl.store.Devices().FindByNamespaceAndDeviceID("default", deviceIDAndURI[0])
	if err != nil && err == storage.ErrNotFound {
		return 0, nil, proto.NewRegistrationError(proto.ErrReasonNoSuchRelam,
			fmt.Sprintf("realm '%s' is not registered", realm))
	} else if err != nil {
		log.Errorf("controller failed to find device: %v", err)
		return 0, nil, proto.NewTechnicalExceptionError(err.Error())
	}

	// Check if session exists
	// TODO(DGL) Fix hardcoded namespace
	existingSess, err := ctrl.store.Sessions().FindByNamespaceAndDeviceID("default", deviceIDAndURI[0])
	if err != nil && err != storage.ErrNotFound {
		log.Errorf("controller failed to search for existing session: %v", err)
		return 0, nil, proto.NewTechnicalExceptionError(err.Error())
	}

	// We found an existing entry, let's check the timeout
	if err == nil {
		if existingSess.LastMessageAt.Add(time.Duration(existingSess.Timeout) * time.Second).After(time.Now()) {
			log.Warnf("controller rejected the control channel becuase session for '%s' exists already", deviceIDAndURI[0])
			return 0, nil, proto.NewRegistrationError(proto.ErrReasonSessionExists,
				fmt.Sprintf("a session for '%s' exists already", realm))
		}

		if err := ctrl.store.Sessions().Delete(existingSess.ID); err != nil {
			log.Errorf("controller failed to delete for expired session: %v", err)
			return 0, nil, proto.NewTechnicalExceptionError(err.Error())
		}
	}

	// Create a new session in the store
	// TODO(DGL) Fix hardcoded namespace
	sess := model.Session{
		Namespace:     "default",
		DeviceID:      device.DeviceID,
		DeviceURI:     device.DeviceURI,
		LastMessageAt: time.Now().Round(time.Second).UTC(),
		Timeout:       device.SessionTimeout,
	}
	if err := ctrl.store.Sessions().Create(&sess); err != nil {
		log.Errorf("controller failed to create new session: %v", err)
		return 0, nil, proto.NewTechnicalExceptionError(err.Error())
	}

	// TODO(DGL) Fix hardcoded namespace
	if err := ctrl.publishDeviceStatus("default", sess.DeviceID, "CONNECTED", sess.ID, sess.LastMessageAt); err != nil {
		log.Errorf("controller could not publish device status: %v", err)
	}

	log.Infof("controller added successfully a new control channel session with ID: %d", sess.ID)

	// Tell control channel that the registration is admitted
	cc.AdmitRegistration(sess.ID, device.SessionTimeout, realm)

	// Return the results of the registration to the control channel
	type registrationDetails struct {
		SessionTimeout int    `json:"session_timeout,omitempty"`
		PingInterval   int    `json:"ping_interval,omitempty"`
		PongTimeout    int    `json:"pong_max_wait_time,omitempty"`
		EventsTopic    string `json:"events_topic,omitempty"`
	}

	details := &registrationDetails{
		SessionTimeout: device.SessionTimeout,
		PingInterval:   device.PingInterval,
		PongTimeout:    device.PongTimeout,
		EventsTopic:    device.EventsTopic,
	}
	return sess.ID, details, nil
}

// UnregisterSession removes a session from the connection and session list.
func (ctrl *Controller) UnregisterSession(sessionID int32) {
	sess, err := ctrl.store.Sessions().FindByID(sessionID)
	if err != nil {
		log.Errorf("controller could not find existing session: %v", err)
		return // No session found we leave
	}

	if err := ctrl.store.Sessions().Delete(sessionID); err != nil {
		log.Errorf("controller failed to delete session from store: %v", err)
	}

	// TODO(DGL) Fix hardcoded namespace
	if err := ctrl.publishDeviceStatus("default", sess.DeviceID, "DISCONNECTED", sess.ID, sess.LastMessageAt); err != nil {
		log.Errorf("controller could not publish device status: %v", err)
	}

	log.Infof("controller removed successfully the control channel session with ID: %d", sessionID)
}
