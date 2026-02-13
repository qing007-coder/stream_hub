package config

import (
	"github.com/spf13/viper"
	"stream_hub/pkg/errors"
	"stream_hub/pkg/model/config"
)

func NewCommonConfig() (*config.CommonConfig, error) {
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigName("common")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(config.CommonConfig)

	if err := v.Unmarshal(conf); err != nil {
		return nil, errors.UnmarshalError
	}

	return conf, nil
}

func NewMediaConfig() (*config.MediaConfig, error) {
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigName("media")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(config.MediaConfig)

	if err := v.Unmarshal(conf); err != nil {
		return nil, errors.UnmarshalError
	}

	return conf, nil
}

func NewUserConfig() (*config.UserConfig, error) {
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigName("user")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(config.UserConfig)

	if err := v.Unmarshal(conf); err != nil {
		return nil, errors.UnmarshalError
	}

	return conf, nil
}

func NewLoggerConfig() (*config.LoggerConfig, error) {
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigName("logger")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(config.LoggerConfig)

	if err := v.Unmarshal(conf); err != nil {
		return nil, errors.UnmarshalError
	}

	return conf, nil
}

func NewVideoConfig() (*config.VideoConfig, error) {
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigName("video")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(config.VideoConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, errors.UnmarshalError
	}

	return conf, nil
}

func NewInteractionConfig() (*config.InteractionConfig, error) {
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigName("interaction")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(config.InteractionConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, errors.UnmarshalError
	}

	return conf, nil
}

func NewGatewayConfig() (*config.GatewayConfig, error) {
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigName("gateway")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(config.GatewayConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, errors.UnmarshalError
	}

	return conf, nil
}

func NewSchedulerConfig() (*config.SchedulerConfig, error) {
	v := viper.New()
	v.AddConfigPath("./config/")
	v.SetConfigName("scheduler")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(config.SchedulerConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, errors.UnmarshalError
	}

	return conf, nil
}
