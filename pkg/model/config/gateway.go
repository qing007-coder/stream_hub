package config

type GatewayConfig struct {
	Name    string  `mapstructure:"name"`
	Port    int     `mapstructure:"port"`
	Service Service `mapstructure:"service"`
}

type Service struct {
	InteractionService string `mapstructure:"interaction_service"`
	VideoService       string `mapstructure:"video_service"`
}
