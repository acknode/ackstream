package xstorage

import "errors"

var ErrCfgNotFound = errors.New("no configs was set")
var ErrConnNotFound = errors.New("no connection was initialized")
var ErrEventInvalid = errors.New("event is not valid")
var ErrEventQueryInvalid = errors.New("event query is not valid")
