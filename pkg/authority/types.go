package authority

type Request struct {
	Operation string      `json:"operation,omitempty"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type ReplyStatus int

const (
	ReplyStatusOK ReplyStatus = iota
	ReplyStatusAbort
	ReplyStatusError
)

type Reply struct {
	Status ReplyStatus `json:"status,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

type AbortResult struct {
	Reason  string      `json:"reason,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type ErrorDetails struct {
	Message string `json:"message,omitempty"`
}

type AuthorizeArguments struct {
	Realm string `json:"realm"`
}

type AuthorizeResult struct {
	SessionTimeout int    `json:"session_timeout,omitempty"`
	PingInterval   int    `json:"ping_interval,omitempty"`
	PongTimeout    int    `json:"pong_max_wait_time,omitempty"`
	EventsTopic    string `json:"events_topic,omitempty"`
}
