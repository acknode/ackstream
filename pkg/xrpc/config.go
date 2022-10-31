package xrpc

import "context"

type Configs struct {
	ServerCertsDir      string `json:"server_certs_dir" mapstructure:"ACKSTREAM_XRPC_SERVER_CERTS_DIR"`
	ServerListenAddress string `json:"server_listen_address" mapstructure:"ACKSTREAM_XRPC_SERVER_LISTEN_ADDRESS"`
	ClientCertFile      string `json:"client_cert_dir" mapstructure:"ACKSTREAM_XRPC_CLIENT_CERT_FILE"`
	ClientRemoteAddress string `json:"client_remote_address" mapstructure:"ACKSTREAM_XRPC_CLIENT_REMOTE_ADDRESS"`
}

type ctxkey string

const CTXKEY_CFG ctxkey = "ackstream.xrpc.configs"

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
