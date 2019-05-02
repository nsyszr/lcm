package devicecontrol

import (
	"io"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

// Handler contains all properties to serve the API
type Handler struct {
	// sessions map[net.Conn]*Session
	sync.RWMutex
}

// NewHandler create a new API handler
func NewHandler() *Handler {
	return &Handler{
		// mgr:      mgr,
		// sessions: make(map[net.Conn]*Session),
	}
}

// RegisterRoutes attaches the handlers to the echo web server
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	log.Debug("Register devicecontrol routes")
	api := e.Group("/devicecontrol")
	api.Any("/v1", h.websocketHandler())
}

func (h *Handler) websocketHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, _, _, err := ws.UpgradeHTTP(c.Request(), c.Response())
		if err != nil {
			return err
		}
		defer conn.Close()

		state := ws.StateServerSide
		ch := wsutil.ControlFrameHandler(conn, state)

		r := &wsutil.Reader{
			Source:         conn,
			State:          state,
			CheckUTF8:      true,
			OnIntermediate: ch,
		}
		w := wsutil.NewWriter(conn, state, 0)

		log.Info("websocket connection established")

		// We're entering now the main loop for a clients specific websocket
		// connection. We don't need to spawn a extra goroutine for each client!
		for {
			h, err := r.NextFrame()
			if err != nil {
				log.Errorf("websocket read message error: %v", err)

				// We should not return the error because echo framework
				// doesn't expect an error at this stage. If you return an
				// error you will see hijacked messages on the console.
				return nil
			}

			if h.OpCode.IsControl() {
				log.Info("websocket control frame received")

				// Check for OpClose before handling the control frame. On
				// OpClose the socket was closed by the client. We can exit our
				// handler now.
				if h.OpCode == ws.OpClose {
					log.Info("websocket connection terminated")
					return nil
				}

				// Handle the control frame
				if err = ch(h, r); err != nil {
					log.Errorf("websocket handle control frame error: %v", err)
					return nil
				}
				continue
			}

			w.Reset(conn, state, h.OpCode)

			if _, err = io.Copy(w, r); err == nil {
				err = w.Flush()
			}
			if err != nil {
				log.Errorf("echo error: %s", err)
				return nil
			}
		}

		// Add session to session map
		// h.Lock()
		// defer h.Unlock()
		// h.sessions[conn] = NewSession()

		// Start listening the websocket connection
		// go listen(conn)
	}
}
