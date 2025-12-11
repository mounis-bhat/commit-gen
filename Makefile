.PHONY: build clean run help install install-home

BINARY_NAME=commit-gen
BUILD_DIR=bin
GO=go

help:
	@echo "Available commands:"
	@echo "  make build        - Build the binary to $(BUILD_DIR)/"
	@echo "  make clean        - Remove the build directory"
	@echo "  make run          - Run with: make run API_KEY=your_key"
	@echo "  make install-home - Build and install to ~/bin"

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/commit-gen

clean:
	rm -rf $(BUILD_DIR)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) $(API_KEY)

install-home: build
	mkdir -p $(HOME)/bin && cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/bin/
