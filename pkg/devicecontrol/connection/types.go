package connection

type messageHandler interface {
	Handle(msg interface{}) ([]byte, Flag, error)
}

type messageHandlerFunc func(msg interface{}) ([]byte, Flag, error)

func (f messageHandlerFunc) Handle(msg interface{}) ([]byte, Flag, error) {
	return f(msg)
}
