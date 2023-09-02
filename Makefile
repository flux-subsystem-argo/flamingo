# Constants
CLI_VERSION      := dev
FLAMINGO_VERSION := $(shell cat SERVER_VERSION)
BINARY_NAME      := flamingo
BIN_DIR          := bin
CMD_DIR          := ./cmd/$(BINARY_NAME)/
BUILD_FLAGS      := -ldflags="-s -w -X main.Version=$(CLI_VERSION) -X main.ServerVersion=$(FLAMINGO_VERSION)"
OUTPUT_PATH      := $(BIN_DIR)/$(BINARY_NAME)

build:
	mkdir -p $(BIN_DIR)
	go fmt ./...
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(OUTPUT_PATH) $(CMD_DIR)
