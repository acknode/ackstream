package xstream

import "errors"

var ErrCfgNotFound = errors.New("no configs was set")
var ErrConnNotFound = errors.New("no connection was intialized")
var ErrMsgInvalidEvent = errors.New("could not construct event from message")
