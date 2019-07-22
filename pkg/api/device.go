package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/nsyszr/lcm/pkg/api/resource"
	"github.com/nsyszr/lcm/pkg/storage"
)

func (h *Handler) handleFetchDevices(c echo.Context) error {
	m, err := h.store.Devices().FetchAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, resource.NewDeviceList(m))
}

func (h *Handler) handleGetDeviceByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	m, err := h.store.Devices().FindByID(int32(id))
	if err != nil && err == storage.ErrNotFound {
		return c.JSON(http.StatusNotFound, err)
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, resource.NewDevice(m))
}

func (h *Handler) handleCreateDevice(c echo.Context) error {
	r := &resource.DeviceResource{}
	if err := c.Bind(r); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	m, err := resource.ValidateDevice(r)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err = h.store.Devices().Create(m)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, resource.NewDevice(m))
}

func (h *Handler) handleDeleteDevice(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	_, err = h.store.Devices().FindByID(int32(id))
	if err != nil && err == storage.ErrNotFound {
		return c.JSON(http.StatusNotFound, err)
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = h.store.Devices().Delete(int32(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusNoContent, nil)
}
