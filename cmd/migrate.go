package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Various migration helpers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
		os.Exit(2)
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
