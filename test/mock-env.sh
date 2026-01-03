#!/bin/bash

# Mock DMS Server environment setup

export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_DEFAULT_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

echo "✅ Environment configured for Mock DMS Server"
echo ""
echo "Mock server endpoint: $AWS_ENDPOINT_URL"
echo "Region: $AWS_DEFAULT_REGION"
echo ""
echo "Make sure the mock server is running:"
echo "  cd test/mock-server && go run main.go"
echo ""
echo "Then test with:"
echo "  ./dms-manager list"
echo "  ./dms-manager describe mock-task-1"
echo "  ./dms-manager tui"
