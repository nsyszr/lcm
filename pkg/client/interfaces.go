package client

type Interface interface {
	RequestCommand(id, cmd string, args []byte) ([]byte, error)
}
