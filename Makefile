.PHONY: help build test mock-server mock-start mock-stop

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the dms-manager binary
	@echo "Building dms-manager..."
	@go build -o dms-manager

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
