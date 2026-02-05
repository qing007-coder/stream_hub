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
