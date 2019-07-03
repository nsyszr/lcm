package controlchannel

import (
	"io/ioutil"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	log "github.com/sirupsen/logrus"
)

func (cc *ControlChannel) inboxWorker() {
	state := ws.StateServerSide
	ch := wsutil.ControlFrameHandler(cc.conn, state)

	r := &wsutil.Reader{
		Source:         cc.conn,
		State:          state,
		CheckUTF8:      true,
		OnIntermediate: ch,
	}

	for {
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
		req, err := ioutil.ReadAll(r)
		if err != nil {
			log.Errorf("websocket read error: %v", err)
			return
		}

		// Handle the received data
		_, _, err = cc.HandleMessage(req)
		if err != nil {
			log.Errorf("websocked handle request error: %v", err)
			return
		}
	}
}

func (cc *ControlChannel) outboxWorker() {
	state := ws.StateServerSide
	w := wsutil.NewWriter(cc.conn, state, 0)

	for {
		select {
		case res := <-cc.wsOutboxCh:
			{
				log.Infof("controlchannel has an outbox message with flag(%d): %s", res.Flag, string(res.Data))
				webSocketWrite(cc.conn, w, state, res, cc.wsTerminateCh)
			}
		case <-cc.wsCloseCh:
			{
				log.Info("controlchannel outbox worker received stop signal")
				webSocketCloseGraceful(cc.conn, w, state, cc.wsTerminateCh)
			}
		}
	}
}

func webSocketWrite(conn net.Conn, w *wsutil.Writer, state ws.State, res *Response, terminateCh chan<- struct{}) {
	var err error

	// Setup the writer with proper websocket frame settings.
	// TODO if we start supporting fragmented message we should rethink
	// this step very well. Maybe it's wrong.
	w.Reset(conn, state, ws.OpText)
	if _, err = w.Write(res.Data); err == nil {
		err = w.Flush()
	}
	if err != nil {
		// TODO We should attach this information to the device log perhaps.
		log.Errorf("websocket write error: %s", err)
		return
	}

	if res.Flag == FlagCloseGracefully {
		webSocketCloseGraceful(conn, w, state, terminateCh)
	} else if res.Flag == FlagTerminate {
		close(terminateCh)
	}
}

func webSocketCloseGraceful(conn net.Conn, w *wsutil.Writer, state ws.State, terminateCh chan<- struct{}) {
	log.Info("websocket graceful close initiated")

	w.Reset(conn, state, ws.OpClose)

	// Write empty string
	var err error
	if _, err = w.Write([]byte("")); err == nil {
		err = w.Flush()
	}
	if err != nil {
		// TODO We should attach this information to the device log perhaps.
		log.Errorf("websocket write error: %s", err)
	}

	close(terminateCh)
}
