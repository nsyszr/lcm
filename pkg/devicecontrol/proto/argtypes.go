package proto

type AbortMessageDetails struct {
	Message string `json:"message"`
}

func NewAbortMessageDetails(message string) *AbortMessageDetails {
	return &AbortMessageDetails{
		Message: message,
	}
}
