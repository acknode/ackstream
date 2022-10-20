package configs

import (
	"github.com/spf13/viper"
	"strings"
)

type Configs struct {
	ListenAddress string `json:"listen_address" mapstructure:"ACKSTREAM_EVENTS_LISTEN_ADDRESS"`
	CertsDir      string `json:"certs_dir" mapstructure:"ACKSTREAM_EVENTS_CERTS_DIR"`
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
	provider.SetDefault("ACKSTREAM_EVENTS_LISTEN_ADDRESS", ":8080")

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
