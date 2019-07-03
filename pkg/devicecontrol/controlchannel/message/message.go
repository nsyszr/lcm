package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

//
// TargetType definition
//

type TargetType int

const (
	TargetTypeSystem = iota
	TargetTypeDevice
)

const TargetBroadcast string = "*"

func (t TargetType) String() string {
	return targetTypeToString[t]
}

var targetTypeToString = map[TargetType]string{
	TargetTypeSystem: "SYSTEM",
	TargetTypeDevice: "DEVICE",
}

var stringToTargetType = map[string]TargetType{
	"SYSTEM": TargetTypeSystem,
	"DEVICE": TargetTypeDevice,
}

func (t TargetType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(targetTypeToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil

	// return json.Marshal(sourceTypeToString[t])
}

func (t *TargetType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*t = stringToTargetType[s]
	return nil
}

//
// SourceType definition
//

type SourceType int

const (
	SourceTypeSystem = iota
	SourceTypeDevice
)

func (t SourceType) String() string {
	return sourceTypeToString[t]
}

var sourceTypeToString = map[SourceType]string{
	SourceTypeSystem: "SYSTEM",
	SourceTypeDevice: "DEVICE",
}

var stringToSourceType = map[string]SourceType{
	"SYSTEM": SourceTypeSystem,
	"DEVICE": SourceTypeDevice,
}

func (t SourceType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(sourceTypeToString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil

	// return json.Marshal(sourceTypeToString[t])
}

func (t *SourceType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*t = stringToSourceType[s]
	return nil
}

func SourceTypeFromString(s string) (SourceType, error) {
	switch s {
	case "DEVICE":
		return SourceTypeDevice, nil
	case "SYSTEM":
		return SourceTypeSystem, nil
	}
	return 0, fmt.Errorf("Invalid source type '%s'", s)
}

type ReplyStatus int

const (
	ReplyStatusSuccess = iota
	ReplyStatusError
)

type PublishRequest struct {
	SourceType SourceType  `json:"source_type"`
	SourceID   string      `json:"source_id,omitempty"`
	TargetType TargetType  `json:"target_type"`
	TargetID   string      `json:"target_id,omitempty"`
	Topic      string      `json:"topic"`
	Arguments  interface{} `json:"arguments,omitempty"`
}

type PublishReply struct {
	Status        ReplyStatus `json:"status"`
	PublicationID int32       `json:"publication_id"`
	ErrorReason   string      `json:"error_reason,omitempty"`
	ErrorDetails  interface{} `json:"error_details,omitempty"`
}

type CallRequest struct {
	TargetType TargetType  `json:"target_type"`
	TargetID   string      `json:"target_id,omitempty"`
	Command    string      `json:"command"`
	Arguments  interface{} `json:"arguments,omitempty"`
}

type CallReply struct {
	Status       ReplyStatus `json:"status"`
	Results      interface{} `json:"results"`
	ErrorReason  string      `json:"error_reason,omitempty"`
	ErrorDetails interface{} `json:"error_details,omitempty"`
}

type ControlChannelCallRequest struct {
	Command   string      `json:"command"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type ControlChannelCallReply struct {
	Status       ReplyStatus `json:"status"`
	Results      interface{} `json:"results"`
	ErrorReason  string      `json:"error_reason,omitempty"`
	ErrorDetails interface{} `json:"error_details,omitempty"`
}

type ControlChannelPublishRequest struct {
	Topic     string      `json:"topic"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type ControlChannelPublishReply struct {
	Status        ReplyStatus `json:"status"`
	PublicationID int32       `json:"publication_id"`
	ErrorReason   string      `json:"error_reason,omitempty"`
	ErrorDetails  interface{} `json:"error_details,omitempty"`
}

type EventMessage struct {
	SourceType    SourceType  `json:"source_type"`
	SourceID      string      `json:"source_id,omitempty"`
	PublicationID int32       `json:"publication_id"`
	Timestamp     time.Time   `json:"timestamp"`
	Details       interface{} `json:"details"`
}
