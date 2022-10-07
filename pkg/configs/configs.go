package configs

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/acknode/ackstream/pkg/xstorage"
	"github.com/acknode/ackstream/pkg/xstream"
	"github.com/spf13/viper"
)

type Configs struct {
	Debug    bool
	Version  string `json:"version" mapstructure:"VERSION"`
	XStream  *xstream.Configs
	XStorage *xstorage.Configs
}

type ctxkey string

const CTXKEY ctxkey = "ackstream.configs"

func WithContext(ctx context.Context, cfg *Configs) context.Context {
	return context.WithValue(ctx, CTXKEY, cfg)
}

func FromContext(ctx context.Context) *Configs {
	configs, ok := ctx.Value(CTXKEY).(*Configs)
	if !ok {
		panic(errors.New("no configs was configured"))
	}

	return configs
}

func IsDebug(envkey string) bool {
	return os.Getenv(envkey) == "dev"
}

func New(provider *viper.Viper, override []string) (*Configs, error) {
	configs := Configs{
		Debug: IsDebug("ACKSTREAM_ENV"),
	}

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
	if err := provider.Unmarshal(&configs.XStream); err != nil {
		return nil, err
	}

	// storage
	if err := provider.Unmarshal(&configs.XStorage); err != nil {
		return nil, err
	}

	return &configs, nil
}

func NewProvider(dirs ...string) (*viper.Viper, error) {
	provider := viper.New()

	provider.AutomaticEnv()

	provider.SetConfigName("configs")
	provider.SetConfigType("props")

	for _, dir := range dirs {
		provider.AddConfigPath(dir)
	}

	if err := provider.MergeInConfig(); err != nil {
		// ignore not found files
		if _, notfound := err.(viper.ConfigFileNotFoundError); !notfound {
			return nil, err
		}
	}

	// set default values
	// common
	provider.SetDefault("ACKSTREAM_REGION", "local")
	provider.SetDefault("ACKSTREAM_VERSION", version())

	// pubsub
	// set stream region to global region by default
	provider.SetDefault("ACKSTREAM_STREAM_URI", "nats://127.0.0.1:4222")
	provider.SetDefault("ACKSTREAM_STREAM_NAME", "ackstream")
	provider.SetDefault("ACKSTREAM_STREAM_REGION", provider.Get("ACKSTREAM_REGION"))

	// storage
	provider.SetDefault("ACKSTREAM_STORAGE_HOSTS", []string{"127.0.0.1"})
	provider.SetDefault("ACKSTREAM_STORAGE_KEYSPACE", "ackstream")
	provider.SetDefault("ACKSTREAM_STORAGE_TABLE", "events")
	provider.SetDefault("ACKSTREAM_STORAGE_BUCKET_TEMPLATE", "20060102")

	return provider, nil
}

func version() string {
	if body, err := os.ReadFile(".version"); err == nil {
		return string(body)
	}

	return "v22.2.2"
}
