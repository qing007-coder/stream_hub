package config

type MediaConfig struct {
	Port      int `mapstructure:"port"`
	ChunkSize int `mapstructure:"chunk_size"`
}
