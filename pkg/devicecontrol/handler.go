package devicecontrol

import (
	"io/ioutil"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo"
	nats "github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol/connection"
	log "github.com/sirupsen/logrus"
)

// Handler contains all properties to serve the API
type Handler struct {
	sessions map[int32]*session
	nc       *nats.Conn
	ctrl     *connection.Controller
	sync.RWMutex
}

// NewHandler create a new API handler
func NewHandler(nc *nats.Conn) *Handler {
	return &Handler{
		sessions: make(map[int32]*session),
		nc:       nc,
		ctrl:     connection.NewController(),
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

		// Determinate the instance of the client.
		instanceID := c.Request().Header.Get("X-INSYS-Service-Instance-Id")
		if instanceID == "" {
			instanceID = "default"
		}

		// The websocket connection is established, let's create a session.
		// The Close methods ensures that the session is removed from the
		// session table on exit.
		cc := connection.NewControlChannel(h.ctrl, conn, w)
		defer cc.Close()

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

			// We reveived an operation control frame and handle it before
			// continuation.
			if h.OpCode.IsControl() {
				log.Info("websocket control frame received")

				// Check for OpClose before handling the control frame. On
				// OpClose the socket was closed by the client. We can exit our
				// handler now.
				if h.OpCode == ws.OpClose {
					// TODO we should attach this information to the device
					// log with a timestamp and modify the discconnectedAt date.
					log.Info("websocket connection closed gracefully")
					return nil
				}

				// Handle the control frame
				if err = ch(h, r); err != nil {
					// TODO We should attach this information to the device log perhaps.
					log.Errorf("websocket handles control frame error: %v", err)
					return nil
				}
				continue
			}

			// Read all data from websocket client
			req, err := ioutil.ReadAll(r)
			if err != nil {
				log.Errorf("websocket read error: %v", err)
				return nil
			}

			// Handle the received data
			res, flag, err := cc.HandleMessage(req)
			if err != nil {
				log.Errorf("websocked handle request error: %v", err)
				return nil
			}

			// Respond data to client back
			if res != nil {
				// Setup the writer with proper websocket frame settings.
				// TODO if we start supporting fragmented message we should rethink
				// this step very well. Maybe it's wrong.
				w.Reset(conn, state, h.OpCode)

				if _, err = w.Write(res); err == nil {
					err = w.Flush()
				}
				if err != nil {
					// TODO We should attach this information to the device log perhaps.
					log.Errorf("websocket write error: %s", err)
					return nil
				}
			}

			// Session handler told us to close the connection gracefully.
			// We send an empty string but with OpClose control frame, to tell
			// the client to close the connection. We will receive back a
			// message with OpClose control frame and this will stop handling
			// the websocket connection. See above.
			if flag == connection.FlagCloseGracefully {
				log.Info("websocket graceful close initiated")
				// Setup the writer with OpClose control frame
				w.Reset(conn, state, ws.OpClose)

				// Write empty string
				if _, err = w.Write([]byte("")); err == nil {
					err = w.Flush()
				}
				if err != nil {
					// TODO We should attach this information to the device log perhaps.
					log.Errorf("websocket write error: %s", err)
					return nil
				}

				// We do not return since we receive an OpClose control frame above
				// return nil
			}

			// Session handler told us to terminate the websocket connection.
			// We exit the handler!
			if flag == connection.FlagTerminate {
				log.Info("websocket terminated")
				return nil
			}
		}
	}
}
