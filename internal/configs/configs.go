package configs

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/acknode/ackstream/pkg/pubsub"
	"github.com/acknode/ackstream/storage"
	"github.com/spf13/viper"
)

type Configs struct {
	Version string `json:"version" mapstructure:"ACKSTREAM_VERSION"`
	Region  string `json:"region" mapstructure:"ACKSTREAM_REGION"`
	PubSub  *pubsub.Configs
	Storage *storage.Configs
}

type ctxkey string

const CTXKEY ctxkey = "ackstream.configs"

func FromContext(ctx context.Context) (*Configs, error) {
	configs, ok := ctx.Value(CTXKEY).(*Configs)
	if !ok {
		return nil, errors.New("no configs was configured")
	}

	return configs, nil
}

func NewConfigs(provider *viper.Viper, override []string) (*Configs, error) {
	configs := Configs{}

	// Allow override configs via parameters
	for _, s := range override {
		kv := strings.Split(s, "=")
		provider.Set(kv[0], kv[1])
	}

	// common
	if err := provider.Unmarshal(&configs); err != nil {
		return nil, err
	}

	// pubsub
	if err := provider.Unmarshal(&configs.PubSub); err != nil {
		return nil, err
	}

	// storage
	if err := provider.Unmarshal(&configs.Storage); err != nil {
		return nil, err
	}

	return &configs, nil
}

func NewProvider(dirs ...string) (*viper.Viper, error) {
	provider := viper.New()
	provider.SetConfigName("configs")
	provider.SetConfigType("props")

	provider.SetEnvPrefix("ACKSTREAM")
	provider.AutomaticEnv()

	for _, dir := range dirs {
		provider.AddConfigPath(dir)
	}

	if err := provider.MergeInConfig(); err != nil {
		// ignore not found files
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// set default values
	// common
	provider.SetDefault("ACKSTREAM_REGION", "local")
	provider.SetDefault("ACKSTREAM_VERSION", version())

	// pubsub
	provider.SetDefault("ACKSTREAM_PUBSUB_URI", "nats://127.0.0.1:4222")
	// set stream region to global region by default
	provider.SetDefault("ACKSTREAM_PUBSUB_STREAM_REGION", provider.Get("ACKSTREAM_REGION"))
	provider.SetDefault("ACKSTREAM_PUBSUB_STREAM_NAME", "ackstream")

	// storage
	provider.SetDefault("ACKSTREAM_STORAGE_HOSTS", []string{"127.0.0.1"})
	provider.SetDefault("ACKSTREAM_STORAGE_KEYSPACE", "ackstream")
	provider.SetDefault("ACKSTREAM_STORAGE_TABLE", "events")

	return provider, nil
}

func version() string {
	if body, err := os.ReadFile(".version"); err == nil {
		return string(body)
	}

	return "v22.2.2"
}
