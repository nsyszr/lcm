package controlchannel

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/wsio"
	"github.com/nsyszr/lcm/pkg/devicecontrol/proto"
	log "github.com/sirupsen/logrus"
)

type Status int

const (
	StatusEstablished Status = iota
	StatusRegistered
)

type sessionDetails struct {
	id            int32
	timeout       int
	realm         string
	lastMessageAt time.Time
}

type ControlChannel struct {
	// sync.RWMutex
	ctrl *Controller
	nc   *nats.Conn
	// conn           net.Conn
	status Status

	sessionDetails      *sessionDetails
	sessionDetailsMutex sync.RWMutex

	stopCh       chan bool
	registeredCh chan bool
	pingCh       chan bool
	// realm          string
	// sessionID      int32
	// sessionTimeout int
	// wsTerminateCh  chan<- struct{}
	// wsCloseCh      chan struct{}
	target *wsio.Driver
	// wsOutboxCh     chan *OutboxMessage
	// inboxCh        chan *InboxMessage

	nextRequestIDMutex sync.RWMutex
	nextRequestID      int32

	callResultsMutex sync.RWMutex
	callResults      map[int32]chan<- interface{}

	subCall *nats.Subscription
}

// Close is called when the websocket handler method is exiting, e.g. the
// connection is closed.
func (cc *ControlChannel) Close() {
	log.Debug("controlchannel close method called")

	// Unregister the control channel from the controller
	cc.ctrl.UnregisterSession(cc.sessionDetails.id)

	if cc.subCall != nil {
		cc.subCall.Unsubscribe()
	}

	// Tell our go waitForPingOrClose routines to stop listening for a signal
	cc.stopCh <- true
}

// inboxHandler listen for messages on targets (websocket driver) inbox channel
func (cc *ControlChannel) inboxHandler() {
	for {
		select {
		case msg := <-cc.target.Inbox:
			{
				log.Debugf("controlchannel reveived message: '%s'", string(msg.Data))

				// Unmarshal the message to get the message type for further processing.
				msgType, msg, err := proto.UnmarshalMessage(msg.Data)
				if err != nil {
					log.Errorf("controlchannel received invalid message: %s", err.Error())
					cc.sendTerminate()
					return // We stop handling new inbox messages
				}

				unhandled := false
				switch msgType {
				case proto.MessageTypeHello:
					err = cc.handleMessage(msg, cc.helloHandler())
				case proto.MessageTypeAbort:
					err = cc.handleMessage(msg, cc.abortHandler())
				case proto.MessageTypePing:
					err = cc.handleMessage(msg, cc.ensureRegistered(cc.keepAliveHandler()))
				case proto.MessageTypePublish:
					err = cc.handleMessage(msg, cc.ensureRegistered(cc.eventHandler()))
				case proto.MessageTypeResult:
					err = cc.handleMessage(msg, cc.ensureRegistered(cc.resultHandler()))
				default:
					unhandled = true
				}

				if err != nil {
					log.Errorf("controlchannel failed to handle message: %s", err.Error())
					cc.sendTerminate()
					return // We stop handling new inbox messages
				}

				if unhandled {
					log.Warnf("controlchannel cannot handle message of type: %s", msgType.String())
					// TODO(DGL) Add error details and check what happens if this
					// method retuns an error. Should we terminate for safety???
					cc.sendAbortMessageAndClose(proto.ErrReasonProtocolViolation,
						fmt.Sprintf("cannot handle message of type '%s'", msgType.String()))
					return // We stop handling new inbox messages
				}
			}
		}
	}
}

// AdmitRegistration is called by the controller after successful registration
// (authorization) of the client. This method sets neccessary values for
// running the control channel and starts the keep alive handling in the
// background (waitForPingOrClose).
func (cc *ControlChannel) AdmitRegistration(sessionID int32, timeout int, realm string) {
	cc.status = StatusRegistered
	cc.updateSessionDetails(sessionID, timeout, realm)

	// Start the session timeout timer. If client doesn't send a ping withing
	// given timeout the connection will be closed.
	go cc.waitForPingOrClose()

	// Listen for call requests
	// TODO(DGL) Working with relam is shit !
	deviceIDAndURI := strings.SplitN(realm, "@", 2)
	go cc.subscribe(deviceIDAndURI[0])

	log.Infof("controlchannel registered for device '%s'", realm)
}

func (cc *ControlChannel) updateSessionDetails(id int32, timeout int, realm string) {
	cc.sessionDetailsMutex.Lock()
	cc.sessionDetails.id = id
	cc.sessionDetails.timeout = timeout
	cc.sessionDetails.realm = realm
	cc.sessionDetailsMutex.Unlock()
}

func (cc *ControlChannel) waitForReqistrationOrClose() {
	log.Debug("controlchannel wait for reqistration routine started")
	for {
		select {
		case <-cc.registeredCh:
			log.Debug("controlchannel wait for reqistration routine received registration signal")
			return
		case <-cc.stopCh:
			log.Debug("controlchannel wait for reqistration routine received stop signal")
			return
		case <-time.After(10 * time.Second): // TODO: get timeout from config
			log.Warn("controlchannel wait for reqistration routine time out")
			cc.target.Stop() // Stop the client connection
			return
		}
	}
}

// messageHandler is a tooling for handling incoming messages. It is similar
// to the go http handler implementation. It allows us to create middleware
// handlers, e.g. the ensureRegistered handler.
type messageHandler interface {
	Handle(msg interface{}) error
}

type messageHandlerFunc func(msg interface{}) error

func (f messageHandlerFunc) Handle(msg interface{}) error {
	return f(msg)
}

// handleMessage is the main method that is called by the public HandleMessage
// function. It expects a handler of interface messageHandler. This method is
// similar to the go implementation of http.HandleFunc.
func (cc *ControlChannel) handleMessage(msg interface{}, h messageHandler) error {
	cc.updateSessionLastMessageAt()
	return h.Handle(msg)
}

func (cc *ControlChannel) updateSessionLastMessageAt() {
	var sessID int32
	lastMessageAt := time.Now().Round(time.Second).UTC()

	// Safely access session details
	cc.sessionDetailsMutex.Lock()
	sessID = cc.sessionDetails.id
	cc.sessionDetails.lastMessageAt = lastMessageAt
	cc.sessionDetailsMutex.Unlock()

	go cc.ctrl.UpdateSession(sessID, lastMessageAt)
}

func (cc *ControlChannel) helloHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) error {
		helloMsg, err := proto.MustHelloMessage(msg)
		if err != nil {
			log.Errorf("controlchannel expected a hello message but error: %s", err)
			return cc.sendTerminate()
		}

		// Notify the waitForReqistrationOrClose go routine that we're about to
		// register the connection, otherwise the connection can be closed
		// during registration.
		cc.registeredCh <- true

		sessID, details, err := cc.ctrl.RegisterSession(cc, helloMsg.Realm)
		if err != nil && proto.IsRegistrationError(err) {
			e := err.(*proto.RegistrationError)
			log.Warnf("controlchannel registration rejected for device '%s' with reason: %s",
				helloMsg.Realm, e.Reason.String())
			return cc.sendAbortMessageAndClose(e.Reason, e.Message)
		} else if err != nil {
			log.Errorf("controlchannel registration failed for device '%s' with error: %s",
				helloMsg.Realm, err.Error())
			return cc.sendTerminate()
		}

		return cc.sendWelcomeMessage(sessID, details)
	})
}

func (cc *ControlChannel) waitForPingOrClose() {
	log.Debug("controlchannel wait for ping routine started")
	for {
		select {
		case <-cc.pingCh:
			log.Debug("controlchannel wait for ping routine received ping signal")
			// We do not exit the loop because we reset the timeout only
		case <-cc.stopCh:
			log.Debug("controlchannel wait for ping routine received stop signal")
			return
		case <-time.After(time.Duration(cc.getSessionTimeout()) * time.Second):
			log.Warn("controlchannel wait for ping routine time out")
			cc.target.Stop() // Stop the client connection
			return
		}
	}
}

func (cc *ControlChannel) getSessionTimeout() int {
	cc.sessionDetailsMutex.RLock()
	t := cc.sessionDetails.timeout
	cc.sessionDetailsMutex.RUnlock()
	return t
}

func (cc *ControlChannel) ensureRegistered(next messageHandler) messageHandler {
	return messageHandlerFunc(func(msg interface{}) error {
		if cc.status != StatusRegistered {
			return cc.sendAbortMessageAndClose(proto.ErrReasonInvalidSession,
				"session not registered")
		}
		return next.Handle(msg)
	})
}

func (cc *ControlChannel) abortHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) error {
		log.Warn("controlchannel terminates the session because of client abort message")
		return cc.sendTerminate()
	})
}

func (cc *ControlChannel) keepAliveHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) error {
		// Notify the waitForPingOrClose method that we received a ping,
		// otherwise session timeout occurs and closes the connection.
		go func() {
			cc.pingCh <- true
		}()

		return cc.sendPongMessage()
	})
}

func (cc *ControlChannel) eventHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) error {
		publishMsg, err := proto.MustPublishMessage(msg)
		if err != nil {
			log.Errorf("controlchannel expected a publish message but error: %s", err)
			return cc.sendTerminate()
		}

		req := message.PublishRequest{
			SourceType: message.SourceTypeDevice,
			SourceID:   "test",
			TargetType: message.TargetTypeSystem,
			Topic:      publishMsg.Topic,
			Arguments:  publishMsg.Arguments,
		}

		requestData, err := json.Marshal(req)
		if err != nil {
			log.Errorf("controlchannel failed to marshal publish request: %s", err)
			return cc.sendTerminate()
		}

		// TODO(DGL) remove hardcoded namespace 'default'
		replyMsg, err := cc.nc.Request("iotcore.devicecontrol.v1.default.publish", requestData, 16*time.Second)
		if err != nil {
			log.Errorf("controlchannel failed to request publish: %s", err)
			return cc.sendTerminate()
		}

		rep := message.PublishReply{}
		if err := json.Unmarshal(replyMsg.Data, &rep); err != nil {
			log.Errorf("controlchannel failed to unmarshal publish reply: %s", err)
			return cc.sendTerminate()
		}

		if rep.Status == message.ReplyStatusError {
			// TODO(DGL) Convert to ErrorReason type
			return cc.sendErrorMessage(proto.MessageTypePublish, publishMsg.RequestID, rep.ErrorReason, rep.ErrorDetails)
		}

		return cc.sendPublishedMessage(publishMsg.RequestID, rep.PublicationID)
	})
}

func (cc *ControlChannel) resultHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) error {
		resultMsg, err := proto.MustResultMessage(msg)
		if err != nil {
			log.Errorf("controlchannel expected a result message but error: %s", err)
			return cc.sendTerminate()
		}

		resultCh := cc.popCallResultCh(resultMsg.RequestID)
		if resultCh == nil {
			// TODO(DGL) should we terminate the control channel here?
			log.Warn("controlchannel received error message but cannot find correlated call message.")
			return cc.sendAbortMessageAndClose(proto.ErrReasonProtocolViolation,
				"Could not handle result for given request id. Time out happend or protocol violation.")
		}
		resultCh <- resultMsg

		return nil
		// We do not respond to a result message bectause it's the response
		// to a call message
	})
}

func (cc *ControlChannel) errorHandler() messageHandlerFunc {
	return messageHandlerFunc(func(msg interface{}) error {
		errorMsg, err := proto.MustErrorMessage(msg)
		if err != nil {
			log.Errorf("controlchannel expected a error message but error: %s", err)
			return cc.sendTerminate()
		}

		switch errorMsg.MessageType {
		case proto.MessageTypeCall:
			{
				resultCh := cc.popCallResultCh(errorMsg.RequestID)
				if resultCh == nil {
					// TODO(DGL) should we terminate the control channel here?
					log.Warn("controlchannel received error message but cannot find correlated call or publish message.")
					return cc.sendAbortMessageAndClose(proto.ErrReasonProtocolViolation,
						"Could not handle result for given request id. Time out happend or protocol violation.")
				}
				resultCh <- errorMsg
			}
		default:
			log.Errorf("controlchannel received error message with invalid message type: %d", errorMsg.MessageType)
			return cc.sendAbortMessageAndClose(proto.ErrReasonProtocolViolation,
				"error message contains invalid message type")
		}

		return nil
		// We do not respond to a error message bectause it's the response
		// to a call or publish message
	})
}

func (cc *ControlChannel) sendTerminate() error {
	return cc.sendMessage(wsio.FlagTerminate, nil)
}

func (cc *ControlChannel) sendAbortMessageAndClose(reason proto.ErrorReason, message string) error {
	out, err := proto.MarshalNewAbortMessage(reason.String(),
		proto.NewAbortMessageDetails(message))
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		log.Errorf("could not marshal abort message: %s", err)
		return cc.sendTerminate()
	}

	return cc.sendMessageAndCloseGraceful(out)
}

func (cc *ControlChannel) sendWelcomeMessage(sessionID int32, details interface{}) error {
	out, err := proto.MarshalNewWelcomeMessage(sessionID, details)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		log.Errorf("could not marshal welcome message: %s", err)
		return cc.sendTerminate()
	}

	return cc.sendMessageAndContinue(out)
}

func (cc *ControlChannel) sendPongMessage() error {
	out, err := proto.MarshalNewPongMessage()
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		log.Errorf("could not marshal pong message: %s", err)
		return cc.sendTerminate()
	}

	return cc.sendMessageAndContinue(out)
}

func (cc *ControlChannel) sendErrorMessage(msgType proto.MessageType, requestID int32, reason string, details interface{}) error {
	out, err := proto.MarshalNewErrorMessage(msgType, requestID, reason, details)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		log.Errorf("could not marshal error message: %s", err)
		return cc.sendTerminate()
	}

	return cc.sendMessageAndContinue(out)
}

func (cc *ControlChannel) sendPublishedMessage(requestID, publicationID int32) error {
	out, err := proto.MarshalNewPublishedMessage(requestID, publicationID)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		log.Errorf("could not marshal published message: %s", err)
		return cc.sendTerminate()
	}

	return cc.sendMessageAndContinue(out)
}

func (cc *ControlChannel) sendCallMessage(requestID int32, operation string, arguments interface{}) error {
	out, err := proto.MarshalNewCallMessage(requestID, operation, arguments)
	// This error should happen never! If it happens log an urgent error
	// and terminate the websocket session for safety.
	if err != nil {
		log.Errorf("could not marshal call message: %s", err)
		return err
	}

	// TODO(DGL) handle full chan buffer
	return cc.sendMessageAndContinue(out)
}

func (cc *ControlChannel) sendMessageAndContinue(data []byte) error {
	return cc.sendMessage(wsio.FlagContinue, data)
}

func (cc *ControlChannel) sendMessageAndCloseGraceful(data []byte) error {
	return cc.sendMessage(wsio.FlagCloseGracefully, data)
}

func (cc *ControlChannel) sendMessage(flag wsio.Flag, data []byte) error {
	select {
	case cc.target.Outbox <- wsio.NewOutboxMessage(flag, data):
		return nil
	default:
		// TODO(DGL) Define better errors
		return fmt.Errorf("outbox is full")
	}
}

func (cc *ControlChannel) getNextRequestID() int32 {
	cc.nextRequestIDMutex.Lock()
	requestID := cc.nextRequestID
	cc.nextRequestID++
	cc.nextRequestIDMutex.Unlock()
	return requestID
}

func (cc *ControlChannel) pushCallResultCh(resultCh chan<- interface{}) int32 {
	requestID := cc.getNextRequestID()

	cc.callResultsMutex.Lock()
	cc.callResultsMutex.Unlock()
	cc.callResults[requestID] = resultCh

	return requestID
}

func (cc *ControlChannel) popCallResultCh(requestID int32) chan<- interface{} {
	cc.callResultsMutex.Lock()
	defer cc.callResultsMutex.Unlock()

	resultCh, ok := cc.callResults[requestID]
	if !ok {
		return nil
	}

	delete(cc.callResults, requestID)
	return resultCh
}
