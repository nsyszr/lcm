package cmd

import (
	"github.com/spf13/cobra"
)

// migrateSQLCmd represents the migrate sql command
var migrateSQLCmd = &cobra.Command{
	Use:   "sql <database-url>",
	Short: "Create SQL schemas and apply migration plans",
	Run:   cmdHandler.Migration.MigrateSQL,
}

func init() {
	migrateCmd.AddCommand(migrateSQLCmd)
}
