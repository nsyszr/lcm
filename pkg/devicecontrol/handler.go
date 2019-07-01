package devicecontrol

import (
	"sync"

	"github.com/gobwas/ws"
	"github.com/labstack/echo"
	nats "github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel"
	log "github.com/sirupsen/logrus"
)

// Handler contains all properties to serve the API
type Handler struct {
	sessions map[int32]*session
	nc       *nats.Conn
	mgr      *controlchannel.Manager
	sync.RWMutex
}

// NewHandler create a new API handler
func NewHandler(nc *nats.Conn) *Handler {
	return &Handler{
		sessions: make(map[int32]*session),
		nc:       nc,
		mgr:      controlchannel.NewManager(),
	}
}

// RegisterRoutes attaches the handlers to the echo web server
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	log.Debug("Register devicecontrol routes")
	api := e.Group("/devicecontrol")
	api.Any("/v1", h.controlChannelHandler())
}

func (h *Handler) controlChannelHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, _, _, err := ws.UpgradeHTTP(c.Request(), c.Response())
		if err != nil {
			return err
		}
		defer conn.Close()

		terminateCh := make(chan struct{})
		cc := controlchannel.New(h.mgr, conn, terminateCh)
		defer cc.Close()
		<-terminateCh
		return nil
	}
}
