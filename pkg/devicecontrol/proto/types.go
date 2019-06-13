package proto

type MessageType int

const (
	MessageTypeInvalid   MessageType = 0
	MessageTypeHello     MessageType = 1
	MessageTypeWelcome   MessageType = 2
	MessageTypeAbort     MessageType = 3
	MessageTypePing      MessageType = 4
	MessageTypePong      MessageType = 5
	MessageTypeError     MessageType = 9
	MessageTypeCall      MessageType = 10
	MessageTypeResult    MessageType = 11
	MessageTypePublish   MessageType = 20
	MessageTypePublished MessageType = 21
)

func (msgType MessageType) String() string {
	names := map[MessageType]string{
		MessageTypeHello:     "HELLO",
		MessageTypeWelcome:   "WELCOME",
		MessageTypeAbort:     "ABORT",
		MessageTypePing:      "PING",
		MessageTypePong:      "PONG",
		MessageTypeError:     "ERROR",
		MessageTypeCall:      "CALL",
		MessageTypeResult:    "RESULT",
		MessageTypePublish:   "PUBLISH",
		MessageTypePublished: "PUBLISHED"}

	msgTypeName, ok := names[msgType]
	if !ok {
		return ""
	}

	return msgTypeName
}

type HelloMessage struct {
	Realm   string
	Details interface{}
}

type WelcomeMessage struct {
	SessionID int32
	Details   interface{}
}

type AbortMessage struct {
	Reason  string
	Details interface{}
}

type PingMessage struct {
	Details interface{}
}

type PongMessage struct {
	Details interface{}
}

type CallMessage struct {
	RequestID int32
	Operation string
	Arguments interface{}
}

type ResultMessage struct {
	RequestID int32
	Results   interface{}
}

type ErrorMessage struct {
	MessageType MessageType
	RequestID   int32
	Error       string
	Details     interface{}
}

type PublishMessage struct {
	RequestID int32
	Topic     string
	Arguments interface{}
}

type PublishedMessage struct {
	RequestID     int32
	PublicationID int32
}
