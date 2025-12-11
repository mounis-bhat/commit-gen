.PHONY: build clean run help

BINARY_NAME=git-commit-generator
GO=go

help:
	@echo "Available commands:"
	@echo "  make build       - Build the binary"
	@echo "  make clean       - Remove the binary"
	@echo "  make run         - Run with: make run API_KEY=your_key"
	@echo "  make install     - Build and install to /usr/local/bin"
	@echo "  make to-bin      - Build and move binary to ./bin/"

build:
	$(GO) build -o $(BINARY_NAME) main.go

clean:
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME) $(API_KEY)

install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/

to-bin: build
	mkdir -p bin && mv $(BINARY_NAME) bin/

install-home: build
	mkdir -p $(HOME)/bin && mv $(BINARY_NAME) $(HOME)/bin/

.PHONY: build clean run help install to-bin install-home
