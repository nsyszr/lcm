package connection

import "github.com/gobwas/ws/wsutil"

type Connection struct {
	Controller *Controller
	Writer     *wsutil.Writer
	Session    *Session
}

type Controller struct {
	sessions map[int32]*Session
}

func NewController() *Controller {
	return &Controller{
		sessions: make(map[int32]*Session, 0),
	}
}
