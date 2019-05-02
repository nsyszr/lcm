package cmd

import (
	"github.com/nsyszr/lcm/cmd/server"
	"github.com/spf13/cobra"
)

// servePublicCmd represents the serve public command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves all endpoints (API and login provider)",
	Run:   server.RunServe(),
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
