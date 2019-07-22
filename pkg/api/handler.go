package api

import (
	"github.com/labstack/echo"
	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/storage"
	log "github.com/sirupsen/logrus"
)

// Handler contains all properties to serve the API
type Handler struct {
	nc    *nats.Conn
	store storage.Interface
}

// NewHandler create a new API handler
func NewHandler(nc *nats.Conn, store storage.Interface) *Handler {
	return &Handler{
		nc:    nc,
		store: store,
	}
}

// RegisterRoutes attaches the handlers to the echo web server
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	log.Debug("Register API routes")
	api := e.Group("/api/v1")
	// api.Any("/v1", h.controlChannelHandler())
	api.GET("/devices", h.handleFetchDevices)
	api.POST("/devices", h.handleCreateDevice)
	api.GET("/devices/:id", h.handleGetDeviceByID)
	api.DELETE("/devices/:id", h.handleDeleteDevice)

	api.GET("/sessions", h.handleFetchSessions)

	api.GET("/events", h.handleFetchEvents)

	api.POST("/call/:namespace/:id", h.handleCallRequest)

	api.Any("/realtime-events", h.realtimeEventsHandler())
}
