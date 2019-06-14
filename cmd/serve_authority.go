package cmd

import (
	"github.com/nsyszr/lcm/pkg/cmd/server"
	"github.com/spf13/cobra"
)

// servePublicCmd represents the serve public command
var serveAuthorityCmd = &cobra.Command{
	Use:   "authority",
	Short: "Serve authority instance",
	Run:   server.RunServeAuthority(),
}

func init() {
	serveCmd.AddCommand(serveAuthorityCmd)
}
