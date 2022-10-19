#!/bin/sh

set -e

docker compose -p ackstream -f docker-compose/docker-compose.yaml down
docker volume prune -f
docker compose -p ackstream -f docker-compose/docker-compose.yaml up -d