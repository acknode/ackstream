#!/bin/bash

set -e
go test -bench=.  -cpu 1,2,3 -benchmem -benchtime 30000x -v github.com/acknode/ackstream/app