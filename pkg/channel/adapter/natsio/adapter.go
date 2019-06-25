package natsio

import (
	"fmt"

	nats "github.com/nats-io/nats.go"
)

type natsAdapter struct {
	nc          *nats.Conn
	id          string
	baseSubject string
	stopCh      chan bool
}

func (a *natsAdapter) commandHandler() error {
	subj := fmt.Sprintf("%s.target.%s.command.*", a.baseSubject, a.id)
	if _, err := a.nc.Subscribe(subj, func(msg *nats.Msg) {
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
		a.nc.Publish(msg.Reply, []byte("result"))
	}); err != nil {
		return err
	}

	return nil
}

func (a *natsAdapter) RequestCommand(id string, data []byte) ([]byte, error) {
	return nil, nil
}
