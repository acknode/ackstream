package xstream

import (
	"context"
)

type Configs struct {
	Debug  bool
	Uri    string `json:"uri" mapstructure:"ACKSTREAM_XSTREAM_URI"`
	Region string `json:"region" mapstructure:"ACKSTREAM_XSTREAM_REGION"`
	Name   string `json:"name" mapstructure:"ACKSTREAM_XSTREAM_NAME"`
	Topic  string `json:"topic" mapstructure:"ACKSTREAM_XSTREAM_TOPIC"`

	MaxMsgs        int64 `json:"max_msg" mapstructure:"ACKSTREAM_XSTREAM_MAX_MSGS"`
	MaxBytes       int64 `json:"max_bytes" mapstructure:"ACKSTREAM_XSTREAM_MAX_BYTES"`
	MaxAge         int32 `json:"max_age" mapstructure:"ACKSTREAM_XSTREAM_MAX_AGE"`
	ConsumerPolicy int   `json:"consumer_policy" mapstructure:"ACKSTREAM_XSTREAM_CONSUMER_POLICY"`
}

type ctxkey string

const CTXKEY_CFG ctxkey = "ackstream.xstream.configs"

var (
	CONSUMER_POLICY_ALL = 0
	CONSUMER_POLICY_NEW = 1
)

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
