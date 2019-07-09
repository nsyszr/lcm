package proto

import "fmt"

type ErrorReason string

const ErrReasonProtocolViolation ErrorReason = "ERR_PROTOCOL_VIOLATION"
const ErrReasonInvalidSession ErrorReason = "ERR_INVALID_SESSION"
const ErrReasonTechnicalException ErrorReason = "ERR_TECHNICAL_EXCEPTION"
const ErrReasonNoSuchRelam ErrorReason = "ERR_NO_SUCH_REALM"
const ErrReasonPublishFailed ErrorReason = "ERR_PUBLISH_FAILED"
const ErrReasonSessionExists ErrorReason = "ERR_SESSION_EXISTS"

func (e ErrorReason) String() string {
	return string(e)
}

type RegistrationError struct {
	Reason  ErrorReason
	Message string
}

func NewRegistrationError(reason ErrorReason, message string) error {
	return &RegistrationError{
		Reason:  reason,
		Message: message,
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
	Reason  ErrorReason
	Message string
}

func NewTechnicalExceptionError(message string) error {
	return &TechnicalExcpetionError{
		Reason:  ErrReasonTechnicalException,
		Message: message,
	}
}

func (e *TechnicalExcpetionError) Error() string {
	return fmt.Sprintf("technical exception: reason: %s", e.Reason)
}

func IsTechnicalExcpetionError(e error) bool {
	_, ok := e.(*TechnicalExcpetionError)
	return ok
}
