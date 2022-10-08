package xstream

import "context"

type Configs struct {
	Uri    string `json:"uri" mapstructure:"ACKSTREAM_XSTREAM_URI"`
	Region string `json:"region" mapstructure:"ACKSTREAM_XSTREAM_REGION"`
	Name   string `json:"name" mapstructure:"ACKSTREAM_XSTREAM_NAME"`
	Topic  string `json:"topic" mapstructure:"ACKSTREAM_XSTREAM_TOPIC"`

	MaxMsgs  int64 `json:"max_msg" mapstructure:"ACKSTREAM_XSTREAM_MAX_MSGS"`
	MaxBytes int64 `json:"max_bytes" mapstructure:"ACKSTREAM_XSTREAM_MAX_BYTES"`
	MaxAge   int32 `json:"max_age" mapstructure:"ACKSTREAM_XSTREAM_MAX_AGE"`
}

const CTXKEY_CFG ctxkey = "xstream.cfg"

func CfgWithContext(ctx context.Context, cfg *Configs) context.Context {
	return context.WithValue(ctx, CTXKEY_CFG, cfg)
}

func CfgFromContext(ctx context.Context) (*Configs, bool) {
	cfg, ok := ctx.Value(CTXKEY_CFG).(*Configs)
	return cfg, ok
}
