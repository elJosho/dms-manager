#!/bin/bash

# Integration test script for DMS Manager with LocalStack

set -e

echo "🧪 Starting DMS Manager Integration Tests"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_pattern="$3"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    echo ""
    echo -n "Testing: $test_name... "
    
    if output=$(eval "$test_command" 2>&1); then
        if echo "$output" | grep -q "$expected_pattern"; then
            echo -e "${GREEN}✓ PASS${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            echo -e "${RED}✗ FAIL${NC} (pattern not found)"
            echo "  Expected: $expected_pattern"
            echo "  Got: $output"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        echo -e "${RED}✗ FAIL${NC} (command failed)"
        echo "  Error: $output"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Check if LocalStack is running
echo ""
echo "Checking LocalStack status..."
if ! curl -s http://localhost:4566/_localstack/health > /dev/null 2>&1; then
    echo -e "${RED}✗ LocalStack is not running${NC}"
    echo "Please start LocalStack first:"
    echo "  make test-setup"
    exit 1
fi
echo -e "${GREEN}✓ LocalStack is running${NC}"

# Check if environment is configured
echo ""
echo "Checking environment configuration..."
if [ -z "$AWS_ENDPOINT_URL" ]; then
    echo -e "${YELLOW}⚠ Environment not configured${NC}"
    echo "Setting up environment..."
    export AWS_ENDPOINT_URL=http://localhost:4566
    export AWS_DEFAULT_REGION=us-east-1
    export AWS_ACCESS_KEY_ID=test
    export AWS_SECRET_ACCESS_KEY=test
fi
echo -e "${GREEN}✓ Environment configured${NC}"
echo "  Endpoint: $AWS_ENDPOINT_URL"
echo "  Region: $AWS_DEFAULT_REGION"

# Build the binary
echo ""
echo "Building dms-manager..."
if go build -o dms-manager; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Run tests
echo ""
echo "========================================"
echo "Running Tests"
echo "========================================"

# Test 1: List command
run_test \
    "List all tasks" \
    "./dms-manager list" \
    "test-task-1"

# Test 2: Describe command
run_test \
    "Describe specific task" \
    "./dms-manager describe test-task-1" \
    "test-task-1"

# Test 3: Multiple describe
run_test \
    "Describe multiple tasks" \
    "./dms-manager describe test-task-1 test-task-2" \
    "test-task-2"

# Test 4: Help command
run_test \
    "Show help" \
    "./dms-manager --help" \
    "Available Commands"

# Test 5: List help
run_test \
    "List command help" \
    "./dms-manager list --help" \
    "List all DMS replication tasks"

# Print results
echo ""
echo "========================================"
echo "Test Results"
echo "========================================"
echo "Tests run:    $TESTS_RUN"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
else
    echo -e "Tests failed: $TESTS_FAILED"
fi
echo "========================================"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
