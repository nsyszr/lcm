package controlchannel

import "fmt"

const ErrReasonTechnicalException = "ERR_TECHNICAL_EXCEPTION"
const ErrReasonNoSuchRelam string = "ERR_NO_SUCH_REALM"
const ErrReasonPublishFailed string = "ERR_PUBLISH_FAILED"

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

type TechnicalExcpetionError struct {
	Reason  string
	Details interface{}
}

func NewTechnicalExceptionError(details interface{}) error {
	return &TechnicalExcpetionError{
		Reason:  ErrReasonTechnicalException,
		Details: details,
	}
}

func (e *TechnicalExcpetionError) Error() string {
	return fmt.Sprintf("technical exception: reason: %s", e.Reason)
}

func IsTechnicalExcpetionError(e error) bool {
	_, ok := e.(*TechnicalExcpetionError)
	return ok
}
