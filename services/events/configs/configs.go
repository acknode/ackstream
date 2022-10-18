package configs

import (
	"github.com/spf13/viper"
	"strings"
)

type Configs struct {
	Port int `json:"port" mapstructure:"PORT"`
}

func NewProvider(dirs ...string) (*viper.Viper, error) {
	provider := viper.New()
	provider.AutomaticEnv()
	provider.SetConfigName("configs")
	provider.SetConfigType("props")

	for _, dir := range dirs {
		provider.AddConfigPath(dir)
		if err := provider.MergeInConfig(); err != nil {
			// ignore not found files, otherwise return error
			if _, notfound := err.(viper.ConfigFileNotFoundError); !notfound {
				return nil, err
			}
		}
	}

	// set default values
	// common
	provider.SetDefault("PORT", 8080)

	return provider, nil
}

func New(provider *viper.Viper, sets []string) (*Configs, error) {
	configs := Configs{}

	// Allow override configs via parameters
	for _, s := range sets {
		kv := strings.Split(s, "=")
		provider.Set(kv[0], kv[1])
	}

	// common
	if err := provider.Unmarshal(&configs); err != nil {
		return nil, err
	}

	return &configs, nil
}
