.PHONY: build clean run help install-home build-all

BINARY_NAME=commit-gen
BUILD_DIR=bin
DIST_DIR=dist
GO=go
LDFLAGS=-s -w

help:
	@echo "Available commands:"
	@echo "  make build        - Build the binary for current platform"
	@echo "  make build-all    - Build for all supported platforms"
	@echo "  make clean        - Remove build directories"
	@echo "  make run          - Run with: make run API_KEY=your_key"
	@echo "  make install-home - Build and install to ~/.local/bin"

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/commit-gen

clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) $(API_KEY)

install-home: build
	mkdir -p $(HOME)/.local/bin && cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/

# Cross-compilation targets
build-all: clean
	@mkdir -p $(DIST_DIR)
	@echo "Building for linux/amd64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64/$(BINARY_NAME) ./cmd/commit-gen
	@echo "Building for linux/arm64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64/$(BINARY_NAME) ./cmd/commit-gen
	@echo "Building for darwin/amd64..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64/$(BINARY_NAME) ./cmd/commit-gen
	@echo "Building for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64/$(BINARY_NAME) ./cmd/commit-gen
	@echo "Building for windows/amd64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64/$(BINARY_NAME).exe ./cmd/commit-gen
	@echo "Building for windows/arm64..."
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64/$(BINARY_NAME).exe ./cmd/commit-gen
	@echo "Done! Binaries in $(DIST_DIR)/"
