#!/bin/bash

# Quick test script - starts mock server and runs basic tests

set -e

echo "🚀 DMS Manager Quick Test"
echo "========================="

# Start mock server in background
echo ""
echo "Starting mock DMS server..."
cd test/mock-server
go run main.go > /tmp/mock-dms-server.log 2>&1 &
MOCK_PID=$!
cd ../..

# Give it a moment to start
sleep 2

# Check if server is running
if ! kill -0 $MOCK_PID 2>/dev/null; then
    echo "❌ Mock server failed to start. Check /tmp/mock-dms-server.log"
    exit 1
fi

echo "✅ Mock server running (PID: $MOCK_PID)"

# Set environment
echo ""
echo "Configuring environment..."
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_DEFAULT_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

# Build
echo ""
echo "Building dms-manager..."
go build -o dms-manager

# Run basic tests
echo ""
echo "========================="
echo "Running Basic Tests"
echo "========================="

echo ""
echo "📋 Listing DMS tasks..."
./dms-manager list
EXIT_CODE=$?

# Cleanup
echo ""
echo "Stopping mock server..."
kill $MOCK_PID 2>/dev/null || true

if [ $EXIT_CODE -eq 0 ]; then
    echo ""
    echo "✅ Quick test successful!"
    echo ""
    echo "Next steps:"
    echo "  1. Start mock server: cd test/mock-server && go run main.go &"
    echo "  2. Source environment: source test/mock-env.sh"
    echo "  3. Try commands:"
    echo "     ./dms-manager list"
    echo "     ./dms-manager describe mock-task-1"
    echo "     ./dms-manager start mock-task-2"
    echo "     ./dms-manager tui"
else
    echo ""
    echo "❌ Test failed"
    exit 1
fi
