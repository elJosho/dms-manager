.PHONY: help build test mock-server mock-start mock-stop build-linux build-linux-arm build-macos build-macos-arm build-windows build-windows-arm build-all clean

# Binary name
BINARY_NAME=dms-manager
BUILD_DIR=build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the dms-manager binary for current platform
	@echo "Building dms-manager..."
	@go build -o $(BINARY_NAME)

# Cross-platform builds
build-linux: ## Build for Linux (amd64)
	@echo "Building for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64

build-linux-arm: ## Build for Linux (arm64)
	@echo "Building for Linux arm64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64

build-macos: ## Build for macOS (amd64/Intel)
	@echo "Building for macOS amd64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64

build-macos-arm: ## Build for macOS (arm64/Apple Silicon)
	@echo "Building for macOS arm64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64

build-windows: ## Build for Windows (amd64)
	@echo "Building for Windows amd64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe

build-windows-arm: ## Build for Windows (arm64)
	@echo "Building for Windows arm64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@GOOS=windows GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe
	@echo "✅ All builds complete in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

mock-server: ## Build the mock DMS server
	@echo "Building mock DMS server..."
	@cd test/mock-server && go build -o mock-dms-server

mock-start: ## Start the mock DMS server in background
	@echo "Starting mock DMS server..."
	@cd test/mock-server && go run main.go > /dev/null 2>&1 &
	@sleep 1
	@echo "✅ Mock server running on http://localhost:4566"
	@echo "   Set environment with: source test/mock-env.sh"

mock-stop: ## Stop the mock DMS server
	@pkill -f "go run main.go" || pkill -f "mock-dms-server" || true
	@echo "✅ Mock server stopped"

test: build ## Run quick test against mock server
	@echo "Testing dms-manager..."
	@echo "Make sure mock server is running (make mock-start)"
	@echo ""
	@source test/mock-env.sh && ./dms-manager list

