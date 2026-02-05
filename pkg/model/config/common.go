package config

type CommonConfig struct {
	Mysql         Mysql         `mapstructure:"mysql"`
	Redis         Redis         `mapstructure:"redis"`
	JWT           JWT           `mapstructure:"jwt"`
	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	SecretKey     string        `mapstructure:"secretKey"`
	Logger        Logger        `mapstructure:"logger.yaml"`
	Email         Email         `mapstructure:"email"`
	Consul        Consul        `mapstructure:"consul"`
	Kafka         Kafka         `mapstructure:"kafka"`
	Clickhouse    Clickhouse    `mapstructure:"clickhouse"`
	Minio         Minio         `mapstructure:"minio"`
	MongoDB       MongoDB       `mapstructure:"mongodb"`
}

type Mysql struct {
	Addr     string `mapstructure:"address"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Name     string `mapstructure:"name"`
	Conf     string `mapstructure:"conf"`
}

type Redis struct {
	Addr     string `mapstructure:"address"`
	Port     string `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

type JWT struct {
	Refresh int `mapstructure:"refresh"`
	Access  int `mapstructure:"access"`
}

type Elasticsearch struct {
	Addr string `mapstructure:"address"`
	Port string `mapstructure:"port"`
}

type Logger struct {
	MaxSize  int `mapstructure:"max_size"`
	Duration int `mapstructure:"duration"`
}

type Email struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
}

type Consul struct {
	Addr string `mapstructure:"addr"`
	Port string `mapstructure:"port"`
}

type Kafka struct {
	Addr string `mapstructure:"addr"`
	Port string `mapstructure:"port"`
}

type Clickhouse struct {
	Addr     string `mapstructure:"addr"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
}

type Minio struct {
	Endpoint  string `mapstructure:"endpoint"`
	Port      string `mapstructure:"port"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

type MongoDB struct {
	Addr string `mapstructure:"addr"`
	Port string `mapstructure:"port"`
}
