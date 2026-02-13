package config

type InteractionConfig struct {
	Name              string `yaml:"name"`
	Port              int    `yaml:"port"`
	SendEventDuration int    `mapstructure:"send_event_duration"`
	MaxEvent          int    `mapstructure:"max_event"`
}
