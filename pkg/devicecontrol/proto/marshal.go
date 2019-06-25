package proto

import (
	"encoding/json"
	"fmt"
)

func (m HelloMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 3)
	envelope[0] = int(MessageTypeHello)
	envelope[1] = m.Realm
	envelope[2] = ensureEmptyDictIfNil(m.Details)

	return json.Marshal(envelope)
}

func (m WelcomeMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 3)
	envelope[0] = int(MessageTypeWelcome)
	envelope[1] = m.SessionID
	envelope[2] = ensureEmptyDictIfNil(m.Details)

	return json.Marshal(envelope)
}

func (m AbortMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 3)
	envelope[0] = int(MessageTypeAbort)
	envelope[1] = m.Reason
	envelope[2] = ensureEmptyDictIfNil(m.Details)

	return json.Marshal(envelope)
}

func (m PingMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 2)
	envelope[0] = int(MessageTypePing)
	envelope[1] = ensureEmptyDictIfNil(m.Details)

	return json.Marshal(envelope)
}

func (m PongMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 2)
	envelope[0] = int(MessageTypePong)
	envelope[1] = ensureEmptyDictIfNil(m.Details)

	return json.Marshal(envelope)
}

func (m CallMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 4)
	envelope[0] = int(MessageTypeCall)
	envelope[1] = m.RequestID
	envelope[2] = m.Operation
	envelope[3] = ensureEmptyDictIfNil(m.Arguments)

	return json.Marshal(envelope)
}

func (m ResultMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 3)
	envelope[0] = int(MessageTypeResult)
	envelope[1] = m.RequestID
	envelope[2] = ensureEmptyDictIfNil(m.Results)

	return json.Marshal(envelope)
}

func (m ErrorMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 5)
	envelope[0] = int(MessageTypeError)
	envelope[1] = int(m.MessageType)
	envelope[2] = m.RequestID
	envelope[3] = m.Error
	envelope[4] = ensureEmptyDictIfNil(m.Details)

	return json.Marshal(envelope)
}

func (m PublishMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 4)
	envelope[0] = int(MessageTypePublish)
	envelope[1] = m.RequestID
	envelope[2] = m.Topic
	envelope[3] = ensureEmptyDictIfNil(m.Arguments)

	return json.Marshal(envelope)
}

func (m PublishedMessage) Marshal() ([]byte, error) {
	envelope := make([]interface{}, 3)
	envelope[0] = int(MessageTypePublished)
	envelope[1] = m.RequestID
	envelope[2] = m.PublicationID

	return json.Marshal(envelope)
}

func MarshalMessage(v interface{}) ([]byte, error) {
	if msg, ok := v.(HelloMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(HelloMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(WelcomeMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(AbortMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(PingMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(PongMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(CallMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(ResultMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(ErrorMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(PublishMessage); ok {
		return msg.Marshal()
	}
	if msg, ok := v.(PublishedMessage); ok {
		return msg.Marshal()
	}
	return nil, fmt.Errorf("cannot marshal an invalid message")
}

func ensureEmptyDictIfNil(v interface{}) interface{} {
	type emptyDict struct{}
	if v == nil {
		return emptyDict{}
	}
	return v
}

func MarshalNewAbortMessage(reason string, details interface{}) ([]byte, error) {
	msg := AbortMessage{Reason: reason, Details: details}
	return MarshalMessage(msg)
}

func MarshalNewWelcomeMessage(sessionID int32, details interface{}) ([]byte, error) {
	msg := WelcomeMessage{SessionID: sessionID, Details: details}
	return MarshalMessage(msg)
}

func MarshalNewPongMessage() ([]byte, error) {
	msg := PongMessage{}
	return MarshalMessage(msg)
}
