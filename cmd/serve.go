package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// servePublicCmd represents the serve public command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Parent command for starting the server instances",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
		os.Exit(2)
	},
	// Run:   server.RunServeDeviceControl(),
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
