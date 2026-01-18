.PHONY: build install clean test lint

BINARY=vip-switch
BUILD_DIR=build
GO=go

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY) ./cmd/vip-switch

install: build
	@echo "Installing $(BINARY)..."
	@mkdir -p /usr/local/bin
	@cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/
	@chmod +x /usr/local/bin/$(BINARY)

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	$(GO) test -v ./...

lint:
	@echo "Running linter..."
	which golangci-lint >/dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run

deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

mod-tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy
