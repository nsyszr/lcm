package authority

import (
	"encoding/json"
	"fmt"
	"time"

	nats "github.com/nats-io/nats.go"
)

type AuthorizeError struct {
	Reason  string
	Details interface{}
}

func NewAuthorizeError(reason string, details interface{}) error {
	return &AuthorizeError{
		Reason:  reason,
		Details: details,
	}
}

func (e *AuthorizeError) Error() string {
	return fmt.Sprintf("autorization failed, reason: %s", e.Reason)
}

func IsAuthorizationError(e error) bool {
	_, ok := e.(*AuthorizeError)
	return ok
}

type AuthorityClient struct {
	nc *nats.Conn
}

func NewAuthorityClient(nc *nats.Conn) *AuthorityClient {
	return &AuthorityClient{
		nc: nc,
	}
}

func (c *AuthorityClient) Authorize(realm string) (*AuthorizeResult, error) {
	// Request
	req := Request{
		Arguments: &AuthorizeArguments{
			Realm: realm,
		},
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msg, err := c.nc.Request("iotcore.devicecontrol.v1.authorize", data, time.Second*10)
	if err != nil {
		return nil, err
	}

	// Response
	reply := Reply{}
	if err := json.Unmarshal(msg.Data, &reply); err != nil {
		return nil, err
	}

	switch reply.Status {
	case ReplyStatusOK:
		// Rerun Unmarshal with the proper Result type
		authResult := &AuthorizeResult{}
		reply := Reply{Result: authResult}
		if err := json.Unmarshal(msg.Data, &reply); err != nil {
			return nil, err
		}
		return authResult, nil
	case ReplyStatusAbort:
		// Rerun Unmarshal with the proper Result type
		abortResult := &AbortResult{}
		reply := Reply{Result: abortResult}
		if err := json.Unmarshal(msg.Data, &reply); err != nil {
			return nil, err
		}
		return nil, NewAuthorizeError(abortResult.Reason, abortResult.Details)
	}
	return nil, fmt.Errorf("unexpected reply for authorization request")
}
