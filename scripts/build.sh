#!/bin/bash

# Check if ML_BUILD_VERSION is set
[ "${ML_BUILD_VERSION}" ] || { echo "ML_BUILD_VERSION is not set"; exit 1; }

# Detect build OS
BUILD_OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Detect build architecture
BUILD_ARCH=$(uname -m)

# Display compilation information
echo "Compiling app for ${BUILD_OS} (${BUILD_ARCH})..."

# Run go mod tidy
go mod tidy

# Build the application
CGO_ENABLED=0 GOOS="${BUILD_OS}" GOARCH="${BUILD_ARCH}" go build \
  -v -trimpath \
  -ldflags="-s -w -X 'github.com/devusSs/minio-link/cmd.BuildVersion=${ML_BUILD_VERSION}' -X 'github.com/devusSs/minio-link/cmd.BuildDate=$(date)' -X 'github.com/devusSs/minio-link/cmd.BuildGitCommit=$(git rev-parse HEAD)'" \
  -o "./.release/minio-link_${BUILD_OS}_${BUILD_ARCH}/" ./...

# Display completion message
echo "Done building app"