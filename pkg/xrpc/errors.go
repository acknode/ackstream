package xrpc

import "errors"

var ErrCfgNotFound = errors.New("no configs was set")
var ErrCACertNotLoad = errors.New("could not load CA certificate")
