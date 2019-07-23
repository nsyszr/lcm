package cli

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	colorable "github.com/mattn/go-colorable"
	"github.com/nsyszr/lcm/config"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type MigrateHandler struct {
	c *config.Config
}

func newMigrateHandler(c *config.Config) *MigrateHandler {
	return &MigrateHandler{c: c}
}

func getDatabaseURL(cmd *cobra.Command, args []string, position int) (url string) {
	if len(args) <= position {
		fmt.Println(cmd.UsageString())
		return
	}
	url = args[position]

	if url == "" {
		fmt.Println(cmd.UsageString())
		return
	}
	return
}

func (h *MigrateHandler) MigrateSQL(cmd *cobra.Command, args []string) {
	url := getDatabaseURL(cmd, args, 0)
	if url == "" {
		os.Exit(2) // Return missing keyword or command
	}

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(colorable.NewColorableStdout())

	log.Info("Applying SQL migration...")

	// Connect to PostgreSQL database
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		log.Errorf("An error occurred while connecting to SQL: %s", err)
		os.Exit(1)
	}
	defer db.Close()

	// Check the database connection
	if err := db.Ping(); err != nil {
		log.Errorf("An error occurred while connecting to SQL: %s", err)
		os.Exit(1)
	}

	// Init db migrations
	migrations := &migrate.FileMigrationSource{
		Dir: "db/migrations",
	}

	// Exec db migrations
	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Errorf("An error occurred while running the migrations: %s", err)
		os.Exit(1)
	}
	log.Infof("Migration successful! Applied a total of %d migrations.", n)
}
