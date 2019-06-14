package authority

import (
	"encoding/json"
	"fmt"

	nats "github.com/nats-io/nats.go"
)

type AuthorityHandler struct {
	nc *nats.Conn
}

func NewAuthorityHandler(nc *nats.Conn) *AuthorityHandler {
	return &AuthorityHandler{
		nc: nc,
	}
}

func (h *AuthorityHandler) Subscribe() error {
	if h.nc == nil {
		return fmt.Errorf("connection to nats is missing")
	}

	if _, err := h.nc.Subscribe("iotcore.devicecontrol.v1.authorize", func(msg *nats.Msg) {
		data, err := h.handleAuthorizeRequest(msg.Data)
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
		h.nc.Publish(msg.Reply, data)
	}); err != nil {
		return err
	}

	return nil
}

func (h *AuthorityHandler) handleAuthorizeRequest(data []byte) ([]byte, error) {
	args := &AuthorizeArguments{}
	req := Request{Arguments: args}

	if err := json.Unmarshal(data, &req); err != nil {
		// Results into a technical exception error
		return nil, err
	}

	if args.Realm != "test" {
		reply := Reply{
			Status: ReplyStatusAbort,
			Result: &AbortResult{
				Reason: "ERR_NO_SUCH_REALM",
			},
		}
		return json.Marshal(reply)
	}

	reply := Reply{
		Status: ReplyStatusOK,
		Result: &AuthorizeResult{
			SessionTimeout: 600,
			PingInterval:   530,
			PongTimeout:    30,
			EventsTopic:    "iotcore.devicecontrol.events",
		},
	}
	return json.Marshal(reply)
}
