package controlchannel

import "fmt"

const ErrReasonNoSuchRelam string = "ERR_NO_SUCH_REALM"

type RegistrationError struct {
	Reason  string
	Details interface{}
}

func NewRegistrationError(reason string, details interface{}) error {
	return &RegistrationError{
		Reason:  reason,
		Details: details,
	}
}

func (e *RegistrationError) Error() string {
	return fmt.Sprintf("registration failed: reason: %s", e.Reason)
}

func IsRegistrationError(e error) bool {
	_, ok := e.(*RegistrationError)
	return ok
}
