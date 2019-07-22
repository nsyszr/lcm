package api

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/nsyszr/lcm/pkg/api/resource"
)

func (h *Handler) handleFetchSessions(c echo.Context) error {
	m, err := h.store.Sessions().FetchAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, resource.NewSessionList(m))
}
