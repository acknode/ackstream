#!/bin/sh

set -e

SERVICES="events"

ORIGINAL_DIR=${PWD}
TEMP_DIR="/tmp/googleapis"
if [ -d $TEMP_DIR ]; then
  cd $TEMP_DIR && git pull && cd "$ORIGINAL_DIR"
else
  git clone https://github.com/googleapis/googleapis.git $TEMP_DIR
fi

for service in $SERVICES
do
  echo "PWD: ${PWD}"
  PROTO_PATH="./services/$service/proto"
  protoc -I "$TEMP_DIR" --proto_path="$PROTO_PATH" --go_out=$PROTO_PATH --go_opt=paths=source_relative --go-grpc_out=$PROTO_PATH --go-grpc_opt=paths=source_relative "$PROTO_PATH/$service.proto"
  protoc -I "$TEMP_DIR" --proto_path="$PROTO_PATH" --grpc-gateway_out=$PROTO_PATH --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative "$PROTO_PATH/$service.proto"
  echo "--> $service"
done
