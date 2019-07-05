package controlchannel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
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
		return err
	}

	subj := fmt.Sprintf("iotcore.devicecontrol.v1.%s.events.devicestatus", namespace)
	return ctrl.nc.Publish(subj, data)
}
