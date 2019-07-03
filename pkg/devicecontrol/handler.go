package devicecontrol

import (
	"github.com/gobwas/ws"
	"github.com/labstack/echo"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel"
	log "github.com/sirupsen/logrus"
)

// Handler contains all properties to serve the API
type Handler struct {
	ctrl *controlchannel.Controller
}

// NewHandler create a new API handler
func NewHandler(ctrl *controlchannel.Controller) *Handler {
	return &Handler{
		ctrl: ctrl,
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
		cc := h.ctrl.NewControlChannel(conn, terminateCh)
		defer cc.Close()
		<-terminateCh
		return nil
	}
}
