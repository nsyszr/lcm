package cmd

import (
	"github.com/nsyszr/lcm/pkg/cmd/server"
	"github.com/spf13/cobra"
)

// servePublicCmd represents the serve public command
var serveDeviceControlCmd = &cobra.Command{
	Use:   "devicecontrol",
	Short: "Serve device control instance",
	Run:   server.RunServeDeviceControl(c),
}

func init() {
	serveCmd.AddCommand(serveDeviceControlCmd)
}
