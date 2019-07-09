package controlchannel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
	"github.com/nsyszr/lcm/pkg/model"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (ctrl *Controller) handlePublishRequest(msg *nats.Msg) error {
	log.Debug("controller handles publish request")
	// Extract the namespace
	// TODO(DGL) Replace hardcoded namespace with namespace from subject
	namespace := "default"

	// Extract the publish request
	req := message.PublishRequest{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		// TODO(DGL) This error should not happen! But what should we do?
		log.Debugf("controller failed to unmarshal publish request: %v", err)
		return errors.Wrap(err, "failed to unmarshal publish request")
	}

	// We received an event which targets the controller. We store this event
	// in our events storage.
	if req.TargetType == message.TargetTypeSystem {

		type errorDetails struct {
			Message string `json:"message"`
		}

		m, err := ctrl.createEventFromPublishRequest(namespace, req)
		if err != nil {
			if err := ctrl.replyPublishFailed(msg.Reply, "ERR_TECHNICAL_EXCEPTION", &errorDetails{Message: err.Error()}); err != nil {
				log.Debugf("controller failed to reply publish request: %v", err)
				return errors.Wrap(err, "failed to reply publish request")
			}
		}

		if err := ctrl.publishCreatedEvent(m); err != nil {
			// TODO(DGL) Add error details to response
			if err := ctrl.replyPublishFailed(msg.Reply, "ERR_TECHNICAL_EXCEPTION", &errorDetails{Message: err.Error()}); err != nil {
				log.Debugf("controller failed to reply publish request: %v", err)
				return errors.Wrap(err, "failed to reply publish request")
			}
			log.Debugf("controller failed to reply publish request: %v", err)
			return errors.Wrap(err, "failed to publish created event")
		}

		if err := ctrl.replyPublishedSuccessfully(msg.Reply, m.ID); err != nil {
			log.Debugf("controller failed to reply publish request: %v", err)
			return errors.Wrap(err, "failed to reply publish request")
		}

		return nil
	}

	// TODO(DGL) handle if TargetType == message.TargetTypeDevice

	return nil
}

func (ctrl *Controller) createEventFromPublishRequest(namespace string, req message.PublishRequest) (*model.Event, error) {
	// Marshall the given request arguments to a string
	details, err := json.Marshal(req.Arguments)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event details")
	}

	m := &model.Event{
		Namespace:  namespace,
		SourceType: req.SourceType.String(),
		SourceID:   req.SourceID,
		Topic:      req.Topic,
		Timestamp:  time.Now().Round(time.Second).UTC(),
		Details:    string(details),
	}

	if err := ctrl.store.Events().Create(m); err != nil {
		return nil, errors.Wrap(err, "failed to store new event")
	}

	return m, nil
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

func (ctrl *Controller) publishCreatedEvent(m *model.Event) error {
	// Unmarshal the details string back to an interface. Since we're marshalling
	// the events message to the queue, we subscribers receives a proper JSON.
	// Otherwise the details are marshalled as an escaped string.
	var details interface{}
	if err := json.Unmarshal([]byte(m.Details), &details); err != nil {
		// TODO(DGL) This error should not happen! But what should we do?
		return errors.Wrap(err, "failed to unmarshal event details")
	}

	srcType, err := message.SourceTypeFromString(m.SourceType)
	if err != nil {
		return errors.Wrap(err, "invalid event source type")
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
		return errors.Wrap(err, "failed to marshal event message")
	}

	subj := fmt.Sprintf("iotcore.devicecontrol.v1.%s.events.%s", m.Namespace, m.Topic)
	if err := ctrl.nc.Publish(subj, data); err != nil {
		return errors.Wrap(err, "failed to publish event message")
	}

	return nil
}
