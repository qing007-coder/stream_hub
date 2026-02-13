package config

type SchedulerConfig struct {
	WorkerNum         int            `mapstructure:"worker_num"`
	HeartbeatExpiry   int            `mapstructure:"heartbeat_expiry"`
	HeartbeatInterval int            `mapstructure:"heartbeat_interval"`
	Concurrency       int            `mapstructure:"concurrency"`
	RegisterKey       string         `mapstructure:"register_key"`
	DeathKey          string         `mapstructure:"death_key"`
	Queue             map[string]int `mapstructure:"queue"`
	Health            HealthConfig   `mapstructure:"health"`
	Janitor           JanitorConfig  `mapstructure:"janitor"`

	Retry      RetryConfig      `mapstructure:"retry"`
	DeadLetter DeadLetterConfig `mapstructure:"dead_letter"`
}

type HealthConfig struct {
	Threshold         int `mapstructure:"threshold"`
	Duration          int `mapstructure:"duration"`
	BlacklistDuration int `mapstructure:"blacklist_duration"`
	Delay             int `mapstructure:"delay"`
}

type JanitorConfig struct {
	HeartbeatInterval int `mapstructure:"heartbeat_interval"`
}

type RetryConfig struct {
	MaxRetries   int64 `mapstructure:"max_retries"`
	BaseDelayMs  int   `mapstructure:"base_delay_ms"`
	MaxDelayMs   int   `mapstructure:"max_delay_ms"`
	EnableJitter bool  `mapstructure:"enable_jitter"`
}

type DeadLetterConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	QueueKey string `mapstructure:"queue_key"`
}
