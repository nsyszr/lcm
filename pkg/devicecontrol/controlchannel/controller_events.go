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

	return nil
}
