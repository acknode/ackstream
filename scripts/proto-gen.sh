#!/bin/sh

set -e

SERVICES="events"
for service in $SERVICES
do
  echo "PWD: ${PWD}"
  PROTO_PATH="./services/$service/protos"

  #  golang
  protoc --proto_path="$PROTO_PATH" --go_out=$PROTO_PATH --go_opt=paths=source_relative --go-grpc_out=$PROTO_PATH --go-grpc_opt=paths=source_relative "$PROTO_PATH/$service.proto"

  echo "--> $service"
done
