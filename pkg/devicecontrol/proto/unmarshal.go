package proto

import (
	"encoding/json"
	"fmt"
)

func unmarshalMessageType(v interface{}) (MessageType, error) {
	msgTypes := map[int]MessageType{
		1:  MessageTypeHello,
		2:  MessageTypeWelcome,
		3:  MessageTypeAbort,
		4:  MessageTypePing,
		5:  MessageTypePong,
		9:  MessageTypeError,
		10: MessageTypeCall,
		11: MessageTypeResult,
		20: MessageTypePublish,
		21: MessageTypePublished}

	i, ok := v.(float64)
	if !ok {
		return MessageTypeInvalid, fmt.Errorf("devicecontrol: invalid message type given")
	}

	msgType, ok := msgTypes[int(i)]
	if !ok {
		return MessageTypeInvalid, fmt.Errorf("devicecontrol: unknown message type given")
	}

	return msgType, nil
}

func UnmarshalMessage(data []byte) (MessageType, interface{}, error) {
	var envelope []interface{}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return MessageTypeInvalid, nil, fmt.Errorf("devicecontrol: invalid message data: %s", err.Error())
	}

	if len(envelope) < 1 {
		return MessageTypeInvalid, nil, fmt.Errorf("devicecontrol: message does not contain a message type")
	}

	msgType, err := unmarshalMessageType(envelope[0])
	if err != nil {
		return msgType, nil, err
	}

	switch msgType {
	case MessageTypeHello:
		return unmarshalHelloMessage(envelope)
	case MessageTypeWelcome:
		return unmarshalWelcomeMessage(envelope)
	case MessageTypeAbort:
		return unmarshalAbortMessage(envelope)
	case MessageTypePing:
		return unmarshalPingMessage(envelope)
	case MessageTypePong:
		return unmarshalPongMessage(envelope)
	case MessageTypeError:
		return unmarshalErrorMessage(envelope)
	case MessageTypeCall:
		return unmarshalCallMessage(envelope)
	case MessageTypeResult:
		return unmarshalResultMessage(envelope)
	case MessageTypePublish:
		return unmarshalPublishMessage(envelope)
	case MessageTypePublished:
		return unmarshalPublishedMessage(envelope)
	}

	// This return should never be reached
	return MessageTypeInvalid, nil, fmt.Errorf("an unexpected error happend during unmarshalling the message")
}

func unmarshalHelloMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) < 2 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete hello message")
	}

	realm, ok := envelope[1].(string)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("hello message contains invalid realm type")
	}

	var details interface{}
	if len(envelope) == 3 {
		details = envelope[2]
	}

	return MessageTypeHello, HelloMessage{
		Realm:   realm,
		Details: details,
	}, nil
}

func unmarshalWelcomeMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) < 2 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete welcome message")
	}

	sessID, ok := envelope[1].(float64)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("welcome message contains invalid session ID type")
	}

	var details interface{}
	if len(envelope) == 3 {
		details = envelope[2]
	}

	return MessageTypeWelcome, WelcomeMessage{
		SessionID: int32(sessID),
		Details:   details,
	}, nil
}

func unmarshalAbortMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) < 2 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete abort message")
	}

	reason, ok := envelope[1].(string)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("abort message contains invalid reason type")
	}

	var details interface{}
	if len(envelope) == 3 {
		details = envelope[2]
	}

	return MessageTypeAbort, AbortMessage{
		Reason:  reason,
		Details: details,
	}, nil
}

func unmarshalPingMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) < 1 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete ping message")
	}

	var details interface{}
	if len(envelope) == 2 {
		details = envelope[1]
	}

	return MessageTypePing, PingMessage{
		Details: details,
	}, nil
}

func unmarshalPongMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) < 1 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete pong message")
	}

	var details interface{}
	if len(envelope) == 2 {
		details = envelope[1]
	}

	return MessageTypePong, PongMessage{
		Details: details,
	}, nil
}

func unmarshalCallMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) < 3 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete call message")
	}

	reqID, ok := envelope[1].(float64)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("call message contains invalid request ID type")
	}

	op, ok := envelope[2].(string)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("call message contains invalid operation type")
	}

	var args interface{}
	if len(envelope) == 4 {
		args = envelope[3]
	}

	return MessageTypeCall, CallMessage{
		RequestID: int32(reqID),
		Operation: op,
		Arguments: args,
	}, nil
}

func unmarshalResultMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) != 3 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete result message")
	}

	reqID, ok := envelope[1].(float64)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("result message contains invalid request ID type")
	}

	return MessageTypeResult, ResultMessage{
		RequestID: int32(reqID),
		Results:   envelope[2],
	}, nil
}

func unmarshalErrorMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) < 4 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete error message")
	}

	msgType, err := unmarshalMessageType(envelope[1])
	if err != nil {
		return MessageTypeInvalid, nil, fmt.Errorf("error message contains invalid or unknown message type")
	}

	reqID, ok := envelope[2].(float64)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("error message contains invalid request ID type")
	}

	e, ok := envelope[3].(string)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("error message contains invalid error type")
	}

	var details interface{}
	if len(envelope) == 5 {
		details = envelope[4]
	}

	return MessageTypeError, ErrorMessage{
		MessageType: msgType,
		RequestID:   int32(reqID),
		Error:       e,
		Details:     details,
	}, nil
}

func unmarshalPublishMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) != 4 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete publish message")
	}

	reqID, ok := envelope[1].(float64)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("publish message contains invalid request ID type")
	}

	topic, ok := envelope[2].(string)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("publish message contains invalid topic type")
	}

	return MessageTypePublish, PublishMessage{
		RequestID: int32(reqID),
		Topic:     topic,
		Arguments: envelope[3],
	}, nil
}

func unmarshalPublishedMessage(envelope []interface{}) (MessageType, interface{}, error) {
	if len(envelope) != 3 {
		return MessageTypeInvalid, nil, fmt.Errorf("incomplete published message")
	}

	reqID, ok := envelope[1].(float64)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("published message contains invalid request ID type")
	}

	pubID, ok := envelope[2].(float64)
	if !ok {
		return MessageTypeInvalid, nil, fmt.Errorf("published message contains invalid publication ID type")
	}

	return MessageTypePublished, PublishedMessage{
		RequestID:     int32(reqID),
		PublicationID: int32(pubID),
	}, nil
}
