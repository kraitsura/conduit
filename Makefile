.PHONY: build clean install test dev daemon release

# Binary name
BINARY=conduit
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Install location
INSTALL_PATH=/usr/local/bin

# Build flags
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

build:
	@echo "Building conduit..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/conduit
	@echo "Built: bin/$(BINARY)"

clean:
	rm -rf bin/ dist/
	go clean

install: build
	@echo "Installing to $(INSTALL_PATH)..."
	sudo cp bin/$(BINARY) $(INSTALL_PATH)/
	@echo "Installed: $(BINARY)"

# Symlink for development (updates live as you rebuild)
link: build
	@echo "Symlinking to $(INSTALL_PATH) for live development..."
	sudo ln -sf $(CURDIR)/bin/$(BINARY) $(INSTALL_PATH)/$(BINARY)
	@echo "Symlinked. Run 'make build' to update."

uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY)

test:
	go test ./...

# Cross-platform release builds
release: clean
	@echo "Building release binaries..."
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/conduit-darwin-arm64 ./cmd/conduit
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/conduit-darwin-amd64 ./cmd/conduit
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/conduit-linux-amd64 ./cmd/conduit
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/conduit-linux-arm64 ./cmd/conduit
	@echo "Release binaries in dist/"
	@ls -la dist/

# Development helpers
dev: build
	./bin/$(BINARY)

daemon: build
	./bin/$(BINARY) daemon

daemon-start: build
	./bin/$(BINARY) daemon start

status: build
	./bin/$(BINARY) status

projects: build
	./bin/$(BINARY) projects
