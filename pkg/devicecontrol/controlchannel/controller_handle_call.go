package controlchannel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
	"github.com/pkg/errors"
)

func (ctrl *Controller) handleCallRequest(msg *nats.Msg) error {
	// Extract the namespace
	// TODO(DGL) Replace hardcoded namespace with namespace from subject
	namespace := "default"

	// Extract the publish request
	req := message.CallRequest{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		// TODO(DGL) This error should not happen! But what should we do?
		return errors.Wrap(err, "failed to unmarshal call request")
	}

	if req.TargetType == message.TargetTypeDevice {
		if req.TargetID == "" {
			// TODO(DGL) Add details for the bad request
			return ctrl.replyCallFailed(msg.Reply, "ERR_BAD_REQUEST", nil)
		}

		// Find a device session for device ID equals target ID
		_, err := ctrl.store.Sessions().FindByNamespaceAndDeviceID(namespace, req.TargetID)
		if err != nil {
			// TODO(DGL) Handle session not found differently
			return ctrl.replyCallFailed(msg.Reply, "ERR_INVALID_SESSION", nil)
		}

		callRequest := message.ControlChannelCallRequest{
			Command:   req.Command,
			Arguments: req.Arguments,
		}

		callRequestData, err := json.Marshal(callRequest)
		if err != nil {
			// TODO(DGL) Add details to error reply
			return ctrl.replyCallFailed(msg.Reply, "ERR_TECHNICAL_EXCEPTION", nil)
		}

		subj := fmt.Sprintf("iotcore.devicecontrol.v1.%s.controlchannel.%s.call", namespace, req.TargetID)
		callReplyMsg, err := ctrl.nc.Request(subj, callRequestData, 16*time.Second)
		if err != nil {
			// TODO(DGL) Add details to error reply
			return ctrl.replyCallFailed(msg.Reply, "ERR_TECHNICAL_EXCEPTION", nil)
		}

		callReply := message.ControlChannelCallReply{}
		if err := json.Unmarshal(callReplyMsg.Data, &callReply); err != nil {
			// TODO(DGL) Add details to error reply
			return ctrl.replyCallFailed(msg.Reply, "ERR_TECHNICAL_EXCEPTION", nil)
		}

		if callReply.Status == message.ReplyStatusError {
			return ctrl.replyCallFailed(msg.Reply, callReply.ErrorReason, callReply.ErrorDetails)
		}

		return ctrl.replyCalledSuccesfully(msg.Reply, callReply.Results)
	}

	return nil
}

func (ctrl *Controller) replyCallFailed(replyTo, reason string, details interface{}) error {
	return ctrl.replyMessage(replyTo, message.CallReply{
		Status:       message.ReplyStatusError,
		ErrorReason:  reason,
		ErrorDetails: details,
	})
}

func (ctrl *Controller) replyCalledSuccesfully(replyTo string, results interface{}) error {
	return ctrl.replyMessage(replyTo, message.CallReply{
		Status:  message.ReplyStatusSuccess,
		Results: results,
	})
}
