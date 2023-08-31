# Constants
FLAMINGO_VERSION := v2.8.1
BINARY_NAME      := flamingo
BIN_DIR          := bin
CMD_DIR          := ./cmd/$(BINARY_NAME)/
BUILD_FLAGS      := -ldflags="-s -w -X main.Version=$(FLAMINGO_VERSION)"
OUTPUT_PATH      := $(BIN_DIR)/$(BINARY_NAME)

build:
	mkdir -p $(BIN_DIR)
	go fmt ./...
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(OUTPUT_PATH) $(CMD_DIR)
