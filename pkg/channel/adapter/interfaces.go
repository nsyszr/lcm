package adapter

type Interface interface {
	RequestCommand([]byte) ([]byte, error)
}
