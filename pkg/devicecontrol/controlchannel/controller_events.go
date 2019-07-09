package controlchannel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
	"github.com/pkg/errors"
)

type deviceStatusDetails struct {
	Status        string    `json:"status"`
	SessionID     int32     `json:"session_id"`
	LastMessageAt time.Time `json:"last_message_at"`
}

func (ctrl *Controller) publishDeviceStatus(namespace, deviceID, status string, sessionID int32, lastMessageAt time.Time) error {
	msg := message.EventMessage{
		SourceType: message.SourceTypeDevice,
		SourceID:   deviceID,
		Timestamp:  time.Now().Round(time.Second).UTC(),
		Details: &deviceStatusDetails{
			Status:        status,
			SessionID:     sessionID,
			LastMessageAt: lastMessageAt,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event message")
	}

	subj := fmt.Sprintf("iotcore.devicecontrol.v1.%s.events.devicestatus", namespace)
	if err := ctrl.nc.Publish(subj, data); err != nil {
		return errors.Wrap(err, "failed to publish event message")
	}

	// ctrl.createDeviceStatusEvent(namespace, deviceID, msg.Details)

	return nil
}

/*func (ctrl *Controller) createDeviceStatusEvent(namespace, deviceID string, details interface{}) (*model.Event, error) {
	// Marshall the given request arguments to a string
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal device status details")
	}

	m := &model.Event{
		Namespace:  namespace,
		SourceType: string(message.SourceTypeDevice),
		SourceID:   deviceID,
		Topic:      "devicestatus",
		Timestamp:  time.Now().Round(time.Second).UTC(),
		Details:    string(detailsJSON),
	}

	if err := ctrl.store.Events().Create(m); err != nil {
		return nil, errors.Wrap(err, "failed to store new event")
	}

	return m, nil
}*/
