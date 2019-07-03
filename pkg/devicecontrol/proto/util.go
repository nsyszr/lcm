package proto

import "fmt"

func MustHelloMessage(v interface{}) (*HelloMessage, error) {
	msg, ok := v.(HelloMessage)
	if !ok {
		return nil, fmt.Errorf("not a hello message")
	}

	return &msg, nil
}

func MustPublishMessage(v interface{}) (*PublishMessage, error) {
	msg, ok := v.(PublishMessage)
	if !ok {
		return nil, fmt.Errorf("not a publish message")
	}

	return &msg, nil
}

func MustResultMessage(v interface{}) (*ResultMessage, error) {
	msg, ok := v.(ResultMessage)
	if !ok {
		return nil, fmt.Errorf("not a result message")
	}

	return &msg, nil
}
