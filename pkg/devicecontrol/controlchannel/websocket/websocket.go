package websocket

import (
	"io/ioutil"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	log "github.com/sirupsen/logrus"
)

type Flag int

const (
	FlagContinue Flag = iota
	FlagCloseGracefully
	FlagTerminate
)

type OutboxMessage struct {
	Flag Flag
	Data []byte
}

type InboxMessage struct {
	Data []byte
}

type WebSocketDriver struct {
	conn   net.Conn
	Inbox  chan *InboxMessage
	Outbox chan *OutboxMessage
	// closeGracefulCh <-chan struct{}

	terminateCh    chan<- struct{}
	terminatedOnce sync.Once

	stopCh   chan struct{}
	stopOnce sync.Once

	wg sync.WaitGroup
}

func NewWebSocketDriver(conn net.Conn, terminateCh chan<- struct{}) *WebSocketDriver {
	return &WebSocketDriver{
		conn:        conn,
		Inbox:       make(chan *InboxMessage, 100),
		Outbox:      make(chan *OutboxMessage, 100),
		terminateCh: terminateCh,
		stopCh:      make(chan struct{}),
	}
}

func (driver *WebSocketDriver) Start( /*stopCh <-chan struct{}*/ ) {
	driver.wg.Add(1)
	go driver.inboxHandler()
	driver.wg.Add(1)
	go driver.outboxHandler()
}

func (driver *WebSocketDriver) Close() {
	// log.Debug("websocketdriver enter close")
	driver.wg.Wait()
	log.Debug("websocketdriver closed")
}

func (driver *WebSocketDriver) Stop() {
	log.Debug("websocketdriver stop called")
	driver.safeCloseTerminateChannel()
}

func (driver *WebSocketDriver) closeHandler() {
	log.Debug("websocketdriver closeHandler called")
	defer driver.wg.Done()
	driver.safeCloseTerminateChannel()
	driver.safeCloseStopChannel()
}

func (driver *WebSocketDriver) safeCloseTerminateChannel() {
	driver.terminatedOnce.Do(func() {
		close(driver.terminateCh)
	})
}

func (driver *WebSocketDriver) safeCloseStopChannel() {
	driver.stopOnce.Do(func() {
		close(driver.stopCh)
	})
}

func (driver *WebSocketDriver) inboxHandler() {
	defer driver.closeHandler()
	// defer safeClose(driver.terminateCh) // On exit handler terminate the websocket (see http control channel handler)

	state := ws.StateServerSide
	ch := wsutil.ControlFrameHandler(driver.conn, state)

	r := &wsutil.Reader{
		Source:         driver.conn,
		State:          state,
		CheckUTF8:      true,
		OnIntermediate: ch,
	}

	for {
		log.Debug("websocket waiting for next frame")
		h, err := r.NextFrame()
		if err != nil {
			// TODO We should attach this information to the device perhaps.
			log.Errorf("websocket read message error: %v", err)

			// We should not return the error because echo framework
			// doesn't expect an error at this stage. If you return an
			// error you will see hijacked messages on the console.
			return
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
				return
			}

			// Handle the control frame
			if err = ch(h, r); err != nil {
				// TODO We should attach this information to the device log perhaps.
				log.Errorf("websocket handles control frame error: %v", err)
				return
			}
			continue
		}

		// Read all data from websocket client
		// log.Debug("websocket reads the frame")
		req, err := ioutil.ReadAll(r)
		if err != nil {
			log.Errorf("websocket read error: %v", err)
			return
		}

		// Handle the received data
		// log.Debugf("websocket received frame with payload: '%s'", string(req))
		// _, _, err = cc.HandleMessage(req)
		driver.Inbox <- NewInboxMessage(req)

		// log.Debug("websocket frame send to handler")

		if err != nil {
			log.Errorf("websocked handle request error: %v", err)
			return
		}
	}
}

func (driver *WebSocketDriver) outboxHandler() {
	defer driver.closeHandler()
	// defer safeClose(driver.terminateCh) // On exit handler terminate the websocket (see http control channel handler)

	state := ws.StateServerSide
	w := wsutil.NewWriter(driver.conn, state, 0)

	for {
		select {
		case res := <-driver.Outbox:
			{
				log.Infof("websocket received an outbox message with flag %d: %s", res.Flag, string(res.Data))
				if err := webSocketWriteText(driver.conn, w, state, res.Data); err != nil {
					// TODO We should attach this information to the device log perhaps.
					log.Errorf("websocket terminates because of write error: %s", err.Error())
					return // stop reading outbox if return value is false, this signals the websocket is about to close!
				}

				switch res.Flag {
				case FlagCloseGracefully:
					{
						log.Info("websocket handled outbox message but closes gracefully")
						webSocketCloseGraceful(driver.conn, w, state)
						return
					}
				case FlagTerminate:
					{
						log.Info("websocket handled outbox message but terminates")
						return
					}
				}
			}
		case <-driver.stopCh:
			return
			/*case <-closeGracefulCh:
			{
				log.Info("websocket received a close graceful signal")
				webSocketCloseGraceful(conn, w, state)
				return // stop reading outbox because the websocket is closed
			}*/
		}
	}
}

func webSocketWriteText(conn net.Conn, w *wsutil.Writer, state ws.State, data []byte) error {
	var err error

	// Setup the writer with proper websocket frame settings.
	// TODO if we start supporting fragmented message we should rethink
	// this step very well. Maybe it's wrong.
	// log.Debug("websocket sending frame to client")
	w.Reset(conn, state, ws.OpText)
	if _, err = w.Write(data); err == nil {
		err = w.Flush()
		// log.Debug("websocket send frame to client finished")
	}
	if err != nil {
		return err
	}

	return nil
}

func webSocketCloseGraceful(conn net.Conn, w *wsutil.Writer, state ws.State) error {
	log.Info("websocket graceful close initiated")

	w.Reset(conn, state, ws.OpClose)

	// Write empty string
	var err error
	if _, err = w.Write([]byte("")); err == nil {
		err = w.Flush()
	}
	if err != nil {
		return err
	}

	return nil
}

func NewOutboxMessage(flag Flag, data []byte) *OutboxMessage {
	m := &OutboxMessage{
		Flag: flag,
	}
	if data != nil {
		m.Data = make([]byte, len(data))
		copy(m.Data, data)
	}
	return m
}

func NewInboxMessage(data []byte) *InboxMessage {
	m := &InboxMessage{}
	if data != nil {
		m.Data = make([]byte, len(data))
		copy(m.Data, data)
	}
	return m
}
