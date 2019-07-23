package cli

import "github.com/nsyszr/lcm/config"

type Handler struct {
	Migration *MigrateHandler
}

func NewHandler(c *config.Config) *Handler {
	return &Handler{
		Migration: newMigrateHandler(c),
	}
}
