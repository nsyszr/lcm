package devicecontrol

import (
	"io/ioutil"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

// Handler contains all properties to serve the API
type Handler struct {
	sessions map[int]*session
	sync.RWMutex
}

// NewHandler create a new API handler
func NewHandler() *Handler {
	return &Handler{
		// mgr:      mgr,
		sessions: make(map[int]*session),
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

		// The websocket connection is established, let's create a session.
		// The Close methods ensures that the session is removed from the
		// session table on exit.
		sess := newSession(h)
		defer sess.close()

		// We're entering now the main loop for a clients specific websocket
		// connection. We don't need to spawn a extra goroutine for each client!
		for {
			h, err := r.NextFrame()
			if err != nil {
				// TODO We should attach this information to the device perhaps.
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
					// TODO we should attach this information to the device
					// log with a timestamp and modify the discconnectedAt date.
					log.Info("websocket connection terminated")
					return nil
				}

				// Handle the control frame
				if err = ch(h, r); err != nil {
					// TODO We should attach this information to the device log perhaps.
					log.Errorf("websocket handle control frame error: %v", err)
					return nil
				}
				continue
			}

			// Read all data from websocket client
			req, err := ioutil.ReadAll(r)
			if err != nil {
				log.Errorf("websocked read error: %v", err)
				return nil
			}

			// Handle the received data
			res, quit, err := sess.handle(req)
			if err != nil {
				log.Errorf("websocked handle request error: %v", err)
				return nil
			}

			// We read and handled everything so reset the I/O writer before
			// responding to client.
			// TODO if we start supporting fragmented message we should rethink
			// this step very well. Maybe it's wrong.
			w.Reset(conn, state, h.OpCode)

			// Respond data to client back, do this before cancel
			if res != nil {
				if _, err = w.Write(res); err == nil {
					err = w.Flush()
				}
				if err != nil {
					// TODO We should attach this information to the device log perhaps.
					log.Errorf("websocket write error: %s", err)
					return nil
				}
			}

			// Session handler told us to close the connection
			if quit {
				log.Info("websocket close initiated")
				conn.Close()
				return nil
			}
		}
	}
}
