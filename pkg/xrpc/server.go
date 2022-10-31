package xrpc

import (
	"context"
	"crypto/tls"
	"github.com/acknode/ackstream/pkg/xlogger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"path/filepath"
)

func NewServer(ctx context.Context, opts []grpc.ServerOption) (*grpc.Server, error) {
	opts, err := WithServerLogger(ctx, opts)
	if err != nil {
		return nil, err
	}
	opts, err = WithServerTLS(ctx, opts)
	if err != nil {
		return nil, err
	}

	return grpc.NewServer(opts...), nil
}

func WithServerTLS(ctx context.Context, opts []grpc.ServerOption) ([]grpc.ServerOption, error) {
	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return opts, err
	}
	logger := xlogger.FromContext(ctx).With("pkg", "xrpc", "fn", "xrpc.server")

	if cfg.ServerCertsDir == "" {
		logger.Debugw("not certificate was given, start with unsecure mode")
		return opts, nil
	}

	// Load server's certificate and private key
	certFile := filepath.Join(cfg.ServerCertsDir, "server-cert.pem")
	keyFile := filepath.Join(cfg.ServerCertsDir, "server-key.pem")
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	})
	logger.Debugw("start secure mode", "cert_file", certFile, "key_file", keyFile)

	return append(opts, grpc.Creds(creds)), nil
}

func WithServerLogger(ctx context.Context, opts []grpc.ServerOption) ([]grpc.ServerOption, error) {
	logger := xlogger.FromContext(ctx).With("pkg", "xrpc")

	opts = append(opts,
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			logger.Debugw("handling request", "request.method", info.FullMethod)
			resp, err = handler(ctx, req)
			return
		}),
	)

	opts = append(opts,
		grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			logger.Debugw("handling request", "request.method", info.FullMethod)
			return handler(srv, ss)
		}),
	)
	return opts, nil
}
