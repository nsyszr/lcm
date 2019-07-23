package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nsyszr/lcm/config"
	"github.com/nsyszr/lcm/pkg/cmd/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var c = new(config.Config)
var cmdHandler = cli.NewHandler(c)

var (
	Version   = "dev-master"
	BuildTime = "undefined"
	GitHash   = "undefined"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "barkeeper",
	Short: "Project Barkeeper (Device Control)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
		os.Exit(2)
	},
}

// Execute runs the idaas and is called by main.main()
func Execute() {
	c.BuildTime = BuildTime
	c.BuildVersion = Version
	c.BuildHash = GitHash

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	} else {
		path := absPathify("$HOME")
		if _, err := os.Stat(filepath.Join(path, ".barkeeper.yml")); err != nil {
			_, _ = os.Create(filepath.Join(path, ".barkeeper.yml"))
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName(".barkeeper") // name of config file (without extension)
		viper.AddConfigPath("$HOME")      // adding home directory as first search path
	}
	viper.AutomaticEnv() // read in environment variables that match

	// Fetch settings
	viper.BindEnv("PORT")
	viper.SetDefault("PORT", 8080)

	viper.BindEnv("HOST")
	viper.SetDefault("HOST", "")

	viper.BindEnv("DATABASE_URL")
	viper.SetDefault("DATABASE_URL", "postgres://u4barkeeper:pw4barkeeper@postgres:5432/barkeeper?sslmode=disable")

	viper.BindEnv("NATS_URL")
	viper.SetDefault("NATS_URL", "nats://nats:4222")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf(`Config file not found because "%s"`, err)
		fmt.Println("")
	}

	if err := viper.Unmarshal(c); err != nil {
		log.Fatal(fmt.Sprintf("Could not read config because %s.", err))
	}
}

func absPathify(inPath string) string {
	if strings.HasPrefix(inPath, "$HOME") {
		inPath = userHomeDir() + inPath[5:]
	}

	if strings.HasPrefix(inPath, "$") {
		end := strings.Index(inPath, string(os.PathSeparator))
		inPath = os.Getenv(inPath[1:end]) + inPath[end:]
	}

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	}
	return ""
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
