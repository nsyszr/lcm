package resource

type RealtimeEventResource struct {
	Namespace string      `json:"namespace"`
	Topic     string      `json:"topic"`
	Data      interface{} `json:"data"`
}

func NewRealtimeEvent(namespace, topic string, data interface{}) *RealtimeEventResource {
	return &RealtimeEventResource{
		Namespace: namespace,
		Topic:     topic,
		Data:      data,
	}
}
