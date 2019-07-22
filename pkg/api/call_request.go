package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
	"github.com/nsyszr/lcm/pkg/storage"
)

func (h *Handler) handleCallRequest(c echo.Context) error {
	namespace := c.Param("namespace")
	deviceID := c.Param("id")

	_, err := h.store.Devices().FindByNamespaceAndDeviceID(namespace, deviceID)
	if err != nil && err == storage.ErrNotFound {
		return c.JSON(http.StatusNotFound, err)
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	req := &message.CallRequest{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	// Override these attributes
	req.TargetType = message.TargetTypeDevice
	req.TargetID = deviceID

	data, err := json.Marshal(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	msg, err := h.nc.Request(fmt.Sprintf("iotcore.devicecontrol.v1.%s.call", namespace), data, 16*time.Second)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	rep := message.CallReply{}
	if err := json.Unmarshal(msg.Data, &rep); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, rep)
}
