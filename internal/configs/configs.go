package configs

import (
	"github.com/acknode/ackstream/pkg/xstream"
	"github.com/acknode/ackstream/utils"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type Configs struct {
	Debug   bool
	Version string           `json:"version" mapstructure:"ACKSTREAM_VERSION"`
	XStream *xstream.Configs `json:"xstream"`
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
	provider.SetDefault("ACKSTREAM_REGION", "local")
	provider.SetDefault("ACKSTREAM_VERSION", version())

	// xstream
	provider.SetDefault("ACKSTREAM_XSTREAM_URI", "nats://127.0.0.1:4222")
	provider.SetDefault("ACKSTREAM_XSTREAM_NAME", "ackstream")
	provider.SetDefault("ACKSTREAM_XSTREAM_REGION", provider.Get("ACKSTREAM_REGION"))
	provider.SetDefault("ACKSTREAM_XSTREAM_TOPIC", "events")
	provider.SetDefault("ACKSTREAM_XSTREAM_MAX_MSG", 8192)      // 8 * 1024
	provider.SetDefault("ACKSTREAM_XSTREAM_MAX_BYTES", 8388608) // 8 * 1024 * 1024
	provider.SetDefault("ACKSTREAM_XSTREAM_MAX_AGE", 1)         // hours

	// xstorage
	provider.SetDefault("ACKSTREAM_XSTORAGE_HOSTS", []string{"127.0.0.1"})
	provider.SetDefault("ACKSTREAM_XSTORAGE_KEYSPACE", "ackstream")
	provider.SetDefault("ACKSTREAM_XSTORAGE_TABLE", "events")
	provider.SetDefault("ACKSTREAM_XSTORAGE_BUCKET_TEMPLATE", "20060102")

	return provider, nil
}

func New(provider *viper.Viper, sets []string) (*Configs, error) {
	configs := Configs{
		Debug: utils.IsDebug("ACKSTREAM_ENV"),
	}

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

func version() string {
	if body, err := os.ReadFile(".version"); err == nil {
		return string(body)
	}

	return "v22.2.2"
}
