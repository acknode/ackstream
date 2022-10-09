package xstorage

import "errors"

var ErrCfgNotSet = errors.New("xstorage: configuration was not set")
var ErrConnNotInit = errors.New("xstorage: connection was not initialized")
