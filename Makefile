# DO NOT CHANGE.
BUILD_OS 				:=
ifeq ($(OS),Windows_NT)
	BUILD_OS = windows
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		BUILD_OS = linux
	endif
	ifeq ($(UNAME_S),Darwin)
		BUILD_OS = darwin
	endif
endif

# DO NOT CHANGE.
BUILD_ARCH 				:=
ifeq ($(echo %PROCESSOR_ARCHITECTURE%), "AMD64")
	BUILD_ARCH = amd64
else
	UNAME_M := $(shell uname -m)
	ifeq ($(UNAME_M), x86_64)
		BUILD_ARCH = amd64
	endif
	ifeq ($(UNAME_M), arm64)
		BUILD_ARCH = arm64
	endif
endif

# DO NOT CHANGE.
build:
	@[ "${LINK_BUILD_VERSION}" ] || ( echo "LINK_BUILD_VERSION is not set"; exit 1 )
	@echo "Compiling app for $(BUILD_OS) ($(BUILD_ARCH))..."
	@go mod tidy
	@CGO_ENABLED=0 GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) go build -v -trimpath -ldflags="-s -w -X 'github.com/devusSs/minio-link/cmd.BuildVersion=${LINK_BUILD_VERSION}' -X 'github.com/devusSs/minio-link/cmd.BuildDate=${shell date}' -X 'github.com/devusSs/minio-link/cmd.BuildGitCommit=${shell git rev-parse HEAD}'" -o ./.release/minio-link_$(BUILD_OS)_$(BUILD_ARCH)/ ./...
	@echo "Done building app"

# DO NOT CHANGE.
help: build
	@-mkdir ./.testing
	@cp ./.release/minio-link_$(BUILD_OS)_$(BUILD_ARCH)/minio-link ./.testing/minio-link
	@clear
	@./.testing/minio-link --help

# DO NOT CHANGE.
version: build
	@-mkdir ./.testing
	@cp ./.release/minio-link_$(BUILD_OS)_$(BUILD_ARCH)/minio-link ./.testing/minio-link
	@clear
	@./.testing/minio-link version

# DO NOT CHANGE.
dev-upload: build
	@-mkdir ./.testing
	@cp ./.release/minio-link_$(BUILD_OS)_$(BUILD_ARCH)/minio-link ./.testing/minio-link
	@clear
	@./.testing/minio-link upload ${TEST_UPLOAD_FILE} -c="./.env" -l="./.logs_dev" -d

# DO NOT CHANGE.
dev-private-upload: build
	@-mkdir ./.testing
	@cp ./.release/minio-link_$(BUILD_OS)_$(BUILD_ARCH)/minio-link ./.testing/minio-link
	@clear
	@./.testing/minio-link upload ${TEST_UPLOAD_FILE} -c="./.env" -l="./.logs_dev" -d -p

# DO NOT CHANGE.
dev-download: build
	@-mkdir ./.testing
	@cp ./.release/minio-link_$(BUILD_OS)_$(BUILD_ARCH)/minio-link ./.testing/minio-link
	@clear
	@./.testing/minio-link download ${TEST_DOWNLOAD_LINK} -c="./.env" -l="./.logs_dev" -d

# DO NOT CHANGE.
check:
	@echo "Checking for potential errors, unused vars, etc."
	@echo "This might take a moment, please wait"
	@go mod tidy
	@go fmt ./...
	@go vet ./...
	@goimports -l ./
	@golangci-lint run ./...
	@gocritic check ./...
	@golines -w ./
	@echo "Done, no errors detected"
	@clear

# DO NOT CHANGE.
clean: check
	@rm -rf ./.testing
	@rm -rf ./.release
	@rm -rf ./logs
	@rm -rf ./files
	@rm -rf ./.logs_dev
	@clear