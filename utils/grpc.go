package utils

import (
	"crypto/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"path/filepath"
)

func NewGRPCServer(certsDir string) (*grpc.Server, error) {
	var grpcOps []grpc.ServerOption
	if certsDir != "" {
		tlsCred, err := LoadServerTLSCred(certsDir)
		if err != nil {
			return nil, err
		}
		grpcOps = append(grpcOps, grpc.Creds(tlsCred))
	}
	return grpc.NewServer(grpcOps...), nil
}

func LoadServerTLSCred(dir string) (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	certFile := filepath.Join(dir, "server-cert.pem")
	keyFile := filepath.Join(dir, "server-key.pem")
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}
