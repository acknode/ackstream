package xstream

import "errors"

var ErrCfgNotSet = errors.New("xstream: configuration was not set")
var ErrConnNotInit = errors.New("xstream: connection was not initialized")
var ErrStreamNotInit = errors.New("xstream: stream was not initialized")
