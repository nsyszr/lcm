package controlchannel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
	"github.com/nsyszr/lcm/pkg/model"
	log "github.com/sirupsen/logrus"
)

func (ctrl *Controller) handlePublishRequest(msg *nats.Msg) error {
	// Extract the namespace
	// TODO(DGL) Replace hardcoded namespace with namespace from subject
	namespace := "default"

	// Extract the publish request
	req := message.PublishRequest{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		// TODO(DGL) This error should not happen! But what should we do?
		return err
	}

	// We received an event which targets the controller. We store this event
	// in our events storage.
	if req.TargetType == message.TargetTypeSystem {
		// Marshall the given request arguments to a string
		details, err := json.Marshal(req.Arguments)
		if err != nil {
			// TODO(DGL) Add error details to response
			if err = ctrl.replyPublishFailed(msg.Reply, ErrReasonPublishFailed, nil); err != nil {
				return err
			}
			return err
		}

		m := model.Event{
			Namespace:  namespace,
			SourceType: req.SourceType.String(),
			SourceID:   req.SourceID,
			Topic:      req.Topic,
			Timestamp:  time.Now().Round(time.Second).UTC(),
			Details:    string(details),
		}

		if err := ctrl.store.Events().Create(&m); err != nil {
			// TODO(DGL) Add error details to response
			if err = ctrl.replyPublishFailed(msg.Reply, ErrReasonPublishFailed, nil); err != nil {
				return err
			}
			return err
		}

		if err := ctrl.publishStoredEvent(&m); err != nil {
			log.Errorf("publish event failed: %s", err.Error())
			// TODO(DGL) Add error details to response
			if err = ctrl.replyPublishFailed(msg.Reply, ErrReasonPublishFailed, nil); err != nil {
				return err
			}
			return err
		}

		return ctrl.replyPublishedSuccessfully(msg.Reply, m.ID)
	}

	// TODO(DGL) handle if TargetType == message.TargetTypeDevice

	return nil
}

func (ctrl *Controller) replyPublishedSuccessfully(replyTo string, publicationID int32) error {
	return ctrl.replyMessage(replyTo, message.PublishReply{
		Status:        message.ReplyStatusSuccess,
		PublicationID: publicationID,
	})
}

func (ctrl *Controller) replyPublishFailed(replyTo, reason string, details interface{}) error {
	return ctrl.replyMessage(replyTo, message.PublishReply{
		Status:       message.ReplyStatusError,
		ErrorReason:  reason,
		ErrorDetails: details,
	})
}

func (ctrl *Controller) publishStoredEvent(m *model.Event) error {
	// Unmarshal the details string back to an interface. Since we're marshalling
	// the events message to the queue, we subscribers receives a proper JSON.
	// Otherwise the details are marshalled as an escaped string.
	var details interface{}
	if err := json.Unmarshal([]byte(m.Details), &details); err != nil {
		// TODO(DGL) This error should not happen! But what should we do?
		return err
	}

	srcType, err := message.SourceTypeFromString(m.SourceType)
	if err != nil {
		return err
	}

	msg := message.EventMessage{
		SourceType:    srcType,
		SourceID:      m.SourceID,
		PublicationID: m.ID,
		Timestamp:     m.Timestamp,
		Details:       details,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	subj := fmt.Sprintf("iotcore.devicecontrol.v1.%s.events.%s", m.Namespace, m.Topic)
	return ctrl.nc.Publish(subj, data)
}
