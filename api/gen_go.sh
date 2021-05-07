#!/usr/bin/env bash
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/.."
protoc --go_out=. --go-grpc_out=. -I=./api/ ./api/grpcoin.proto
