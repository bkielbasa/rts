# Tanks RTS Makefile

BINARY_NAME=tanks
BUILD_DIR=build
CMD_GAME=./cmd/game
CMD_GENMAP=./cmd/genmap

# Go build flags
LDFLAGS=-s -w

.PHONY: all build build-mac build-linux build-windows build-all clean run genmap

# Default target
all: build

# Build for current platform
build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_GAME)

# Build for macOS (Intel and Apple Silicon)
build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-mac-amd64 $(CMD_GAME)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-mac-arm64 $(CMD_GAME)

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_GAME)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_GAME)

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_GAME)

# Build for all platforms
build-all: build-mac build-linux build-windows

# Run the game
run:
	go run $(CMD_GAME)

# Generate a new map
genmap:
	go run $(CMD_GENMAP)

# Build the map generator
build-genmap:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/genmap $(CMD_GENMAP)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean
