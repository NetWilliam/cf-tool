#!/bin/bash
# Simple build script that mimics 'make build' for cf-tool

set -e

export PATH=$PATH:~/go/bin
export GOTOOLCHAIN=auto
export GOPROXY=https://goproxy.cn,direct

BINARY_NAME=cf
BINARY_PATH=./bin/${BINARY_NAME}
BUILD_DIR=./bin

echo "Building ${BINARY_NAME} for Linux..."
mkdir -p ${BUILD_DIR}

# Get version info
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build with version info
go build -v -mod=mod -ldflags "-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" -o ${BINARY_PATH} cf.go

echo "Build complete: ${BINARY_PATH}"
echo "Version: ${VERSION}"
echo "Build time: ${BUILD_TIME}"
