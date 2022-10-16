package xstorage

import "context"

type Configs struct {
	Hosts    []string `json:"hosts" mapstructure:"ACKSTREAM_XSTORAGE_HOSTS"`
	Keyspace string   `json:"keyspace" mapstructure:"ACKSTREAM_XSTORAGE_KEYSPACE"`
	Table    string   `json:"table" mapstructure:"ACKSTREAM_XSTORAGE_TABLE"`
}

type ctxkey string

const CTXKEY_CFG ctxkey = "ackstream.xstorage.configs"

func CfgWithContext(ctx context.Context, cfg *Configs) context.Context {
	return context.WithValue(ctx, CTXKEY_CFG, cfg)
}

func CfgFromContext(ctx context.Context) (*Configs, error) {
	configs, ok := ctx.Value(CTXKEY_CFG).(*Configs)
	if !ok {
		return nil, ErrCfgNotFound
	}
	return configs, nil
}
