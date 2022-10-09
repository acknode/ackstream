package xstorage

import "context"

const CTXKEY_CFG ctxkey = "xstorage.cfg"

type Configs struct {
	Hosts          []string `json:"hosts" mapstructure:"ACKSTREAM_XSTORAGE_HOSTS"`
	Keyspace       string   `json:"keyspace" mapstructure:"ACKSTREAM_XSTORAGE_KEYSPACE"`
	Table          string   `json:"table" mapstructure:"ACKSTREAM_XSTORAGE_TABLE"`
	BucketTemplate string   `json:"bucket_template" mapstructure:"ACKSTREAM_XSTORAGE_BUCKET_TEMPLATE"`
}

func CfgWithContext(ctx context.Context, cfg *Configs) context.Context {
	return context.WithValue(ctx, CTXKEY_CFG, cfg)
}

func CfgFromContext(ctx context.Context) (*Configs, bool) {
	cfg, ok := ctx.Value(CTXKEY_CFG).(*Configs)
	return cfg, ok
}
