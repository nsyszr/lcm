package proto

import "fmt"

func MustHelloMessage(v interface{}) (*HelloMessage, error) {
	msg, ok := v.(HelloMessage)
	if !ok {
		return nil, fmt.Errorf("not a hello message")
	}

	return &msg, nil
}
