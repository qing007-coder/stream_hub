package config

type VideoConfig struct {
	Name              string `mapstructure:"name"`
	Port              int    `mapstructure:"port"`
	SendEventDuration int    `mapstructure:"send_event_duration"`
	MaxEvent          int    `mapstructure:"max_event"`
}
