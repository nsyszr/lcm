package config

// Config contains all application settings
type Config struct {
	BindPort      int    `mapstructure:"PORT" yaml:"port"`
	BindHost      string `mapstructure:"HOST" yaml:"host"`
	DatabaseURL   string `mapstructure:"DATABASE_URL" yaml:"database_url"`
	NATSServerURL string `mapstructure:"NATS_URL" yaml:"nats_url"`

	// Version
	BuildVersion string `yaml:"-"`
	BuildHash    string `yaml:"-"`
	BuildTime    string `yaml:"-"`
}
