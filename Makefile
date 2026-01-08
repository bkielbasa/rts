# Tanks RTS Makefile

BINARY_NAME=tanks
BUILD_DIR=build
DIST_DIR=dist
CMD_GAME=./cmd/game
CMD_GENMAP=./cmd/genmap
VERSION?=1.0.0

# Go build flags
LDFLAGS=-s -w

.PHONY: all build build-mac build-linux build-windows build-all clean run genmap dist-windows

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

# Create Windows distribution ZIP with binary and assets
dist-windows: build-windows
	@echo "Creating Windows distribution..."
	@mkdir -p $(DIST_DIR)/$(BINARY_NAME)-windows-$(VERSION)
	@cp $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(DIST_DIR)/$(BINARY_NAME)-windows-$(VERSION)/$(BINARY_NAME).exe
	@cp -r assets $(DIST_DIR)/$(BINARY_NAME)-windows-$(VERSION)/
	@cp -r maps $(DIST_DIR)/$(BINARY_NAME)-windows-$(VERSION)/
	@cd $(DIST_DIR) && zip -r $(BINARY_NAME)-windows-$(VERSION).zip $(BINARY_NAME)-windows-$(VERSION)
	@rm -rf $(DIST_DIR)/$(BINARY_NAME)-windows-$(VERSION)
	@echo "Created: $(DIST_DIR)/$(BINARY_NAME)-windows-$(VERSION).zip"

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
	rm -rf $(DIST_DIR)
	go clean
