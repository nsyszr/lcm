package controlchannel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
	"github.com/nsyszr/lcm/pkg/devicecontrol/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (cc *ControlChannel) subscribe(deviceID string) error {
	if cc.nc == nil {
		// TODO(DGL) Create a true error
		return fmt.Errorf("controlchannel: connection to nats is missing")
	}

	// TODO(DGL) Replace hardcoded namespace and device ID
	subj := fmt.Sprintf("iotcore.devicecontrol.v1.%s.controlchannel.%s.call", "default", deviceID)
	sub, err := cc.nc.Subscribe(subj, func(msg *nats.Msg) {
		log.Debugf("controlchannel received message from call queue: %s", string(msg.Data))

		// Start handling of reply async ! The method will exit always because
		// there's an timeout. This ensures that the subscribe method isn't
		// blocked. Sometimes NATS repeat sending a message.
		go cc.handleCallRequestOrTimeout(msg)
		/*if err := cc.handleCallRequest(msg); err != nil {
			log.Error("controlchannel failed to handle call request: ", err.Error())
		}*/
	})
	if err != nil {
		return errors.Wrap(err, "failed to subscribe the controlchannel call queue")
	}

	cc.subCall = sub

	return nil
}

func (cc *ControlChannel) handleCallRequestOrTimeout(msg *nats.Msg) error {
	log.Debug("controlchannel started handel call request routine")
	// Extract the publish request
	req := message.ControlChannelCallRequest{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		// TODO(DGL) This error should not happen! But what should we do?
		return errors.Wrap(err, "failed to unmarshal controlchannel call request")
	}

	resultCh := make(chan interface{})
	requestID := cc.pushCallResultCh(resultCh)

	if err := cc.sendCallMessage(requestID, req.Command, req.Arguments); err != nil {
		return errors.Wrap(err, "failed to send call message")
	}

	for {
		log.Debug("controlchannel wait for call result")
		select {
		// TODO(DGL) If we set the same timeout of 16 seconds here, we run into
		// problems with the requestor. NATS responds with timeout before this
		// message arrives to the requestor. But in this case the device result
		// response is timed out. We need properly defined settings!
		// TODO(DGL) I think it doesn't make sense for a timeout reply since
		// the request will by timed out by the queue. If we didn't receive
		// a reply from websocket we should terminate the session!
		case <-time.After(16 * time.Second):
			log.Error("controlchannel call request timed out")
			// return cc.replyCallFailed(msg.Reply, "ERR_TIMEOUT", nil)

			// TODO: try to remove resultCh from map. If client sends message
			// later the result handler will response with an error and quits
			// the exisiting session.
			/* return cc.sendAbortMessageAndClose("ERR_PROTOCOL_VIOLATION",
			proto.NewAbortMessageDetails("result message timeout"))*/
			cc.popCallResultCh(requestID)
			return cc.replyCallFailed(msg, "ERR_RESULT_TIMEOUT", nil)
		case result := <-resultCh:
			log.Debug("controlchannel handle call request routine reveived a result")
			resultMsg, ok := result.(*proto.ResultMessage)
			if ok {
				return cc.replyCalledSuccesfully(msg, resultMsg.Results)
			}
			errorMsg, ok := result.(*proto.ErrorMessage)
			if ok {
				return cc.replyCallFailed(msg, errorMsg.Error, errorMsg.Details)
			}
			return cc.replyCallFailed(msg, "ERR_TECHNICAL_EXCEPTION", nil)
		}
	}
}

func (cc *ControlChannel) replyCallFailed(msg *nats.Msg, reason string, details interface{}) error {
	return cc.replyMessage(msg, message.ControlChannelCallReply{
		Status:       message.ReplyStatusError,
		ErrorReason:  reason,
		ErrorDetails: details,
	})
}

func (cc *ControlChannel) replyCalledSuccesfully(msg *nats.Msg, results interface{}) error {
	return cc.replyMessage(msg, message.ControlChannelCallReply{
		Status:  message.ReplyStatusSuccess,
		Results: results,
	})
}

func (cc *ControlChannel) replyMessage(msg *nats.Msg, rep interface{}) error {
	data, err := json.Marshal(rep)
	if err != nil {
		return errors.Wrap(err, "failed to marshal reply message")
	}
	if err := msg.Respond(data); err != nil {
		return errors.Wrap(err, "failed to respond the request message")
	}
	return nil
}
