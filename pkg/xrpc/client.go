package xrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/acknode/ackstream/pkg/xlogger"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
)

func NewClient(ctx context.Context, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	opts, err = WithClientRetry(ctx, opts)
	if err != nil {
		return nil, err
	}

	opts, err = WithClientTLS(ctx, opts)
	if err != nil {
		return nil, err
	}

	return grpc.Dial(cfg.ClientRemoteAddress, opts...)
}

func WithClientTLS(ctx context.Context, opts []grpc.DialOption) ([]grpc.DialOption, error) {
	cfg, err := CfgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	logger := xlogger.FromContext(ctx).With("pkg", "xrpc", "fn", "xrpc.client")

	if cfg.ClientCertsDir == "" {
		logger.Debugw("not certificate was given, start with unsecure mode")
		return append(opts, grpc.WithTransportCredentials(insecure.NewCredentials())), nil
	}

	caFile := filepath.Join(cfg.ServerCertsDir, "ca-cert.pem")
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, ErrCACertNotLoad
	}
	logger.Debugw("start secure mode", "ca_file", caFile)

	// Create the credentials and return it
	opt := grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{RootCAs: certPool}))
	return append(opts, opt), nil
}

func WithClientRetry(ctx context.Context, opts []grpc.DialOption) ([]grpc.DialOption, error) {
	streamOpt := grpcRetry.StreamClientInterceptor(grpcRetry.WithMax(3))
	opts = append(opts, grpc.WithStreamInterceptor(streamOpt))

	unaryOpt := grpcRetry.UnaryClientInterceptor(grpcRetry.WithMax(3))
	opts = append(opts, grpc.WithUnaryInterceptor(unaryOpt))

	return opts, nil
}
