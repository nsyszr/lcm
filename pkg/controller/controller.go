package controller

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

type Controller struct {
	nc *nats.Conn
}

func New(nc *nats.Conn) *Controller {
	return &Controller{
		nc: nc,
	}
}

func (ctrl *Controller) Subscribe() error {
	if ctrl.nc == nil {
		return fmt.Errorf("controller: connection to nats is missing")
	}

	if _, err := ctrl.nc.QueueSubscribe("iotcore.devicecontrol.v1.>", "iotcore.devicecontrol.v1.controllers", func(msg *nats.Msg) {
		ctrl.handleMessage(msg)
		/*data, err := h.handleAuthorizeRequest(msg.Data)
		if err != nil {
			// TODO(DGL) Review this
			// log.xxxx
			reply := Reply{
				Status: ReplyStatusAbort,
				Result: &AbortResult{
					Reason:  "ERR_TECHNICAL_EXCEPTION",
					Details: &ErrorDetails{Message: err.Error()},
				},
			}
			res, _ := json.Marshal(reply)
			h.nc.Publish(msg.Reply, res)
			return
		}
		h.nc.Publish(msg.Reply, data)*/
	}); err != nil {
		return err
	}

	return nil
}

func (ctrl *Controller) handleMessage(msg *nats.Msg) {
	if strings.HasPrefix(msg.Subject, "iotcore.devicecontrol.v1.command") {
		ctrl.handleCommand(msg)
	}
}

func (ctrl *Controller) handleCommand(msg *nats.Msg) error {
	req := CommandRequest{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return err
	}

	// Find the device command queue

	responseMsg, err := ctrl.nc.Request("iotcore.devicecontrol.v1.controlchannel.test.command", msg.Data, 16*time.Second)
	if err != nil {
		return err
	}

	res := &CommandReqly{}
	if err := json.Unmarshal(responseMsg.Data, res); err != nil {
		return err
	}
	res.DeviceID = req.DeviceID

	return ctrl.nc.Publish(msg.Reply, msg.Data)
}
