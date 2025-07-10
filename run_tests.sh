#!/bin/bash

# SmarApp API Test Runner
# This script runs all tests and provides coverage information

set -e

echo "?? SmarApp API Test Suite"
echo "========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean up any existing test databases
print_status "Cleaning up test artifacts..."
find . -name "*.db" -type f -delete 2>/dev/null || true

# Run go mod tidy to ensure dependencies are up to date
print_status "Ensuring dependencies are up to date..."
go mod tidy

# Run tests with coverage
print_status "Running unit tests with coverage..."

# Test individual packages
echo ""
echo "?? Testing Models..."
go test ./models/... -v -cover

echo ""
echo "?? Testing Middleware..."
go test ./middleware/... -v -cover

echo ""
echo "?? Testing Handlers..."
go test ./handlers/... -v -cover

# Run all tests together for overall coverage
echo ""
echo "?? Generating overall coverage report..."
go test ./... -coverprofile=coverage.out -covermode=atomic

# Generate coverage report
if command -v go &> /dev/null; then
    echo ""
    print_status "Coverage Summary:"
    go tool cover -func=coverage.out | tail -1
    
    # Generate HTML coverage report
    print_status "Generating HTML coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    print_success "Coverage report generated: coverage.html"
fi

# Run integration tests (API tests)
echo ""
echo "?? Running Integration Tests..."
print_status "Building application..."
go build -o smarapp-api-test cmd/server/main.go

print_status "Starting test server..."
./smarapp-api-test &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Check if server is running
if kill -0 $SERVER_PID 2>/dev/null; then
    print_success "Test server started (PID: $SERVER_PID)"
    
    # Run API tests
    print_status "Running API integration tests..."
    if ./test_api.sh; then
        print_success "API integration tests passed!"
    else
        print_error "API integration tests failed!"
    fi
    
    # Stop the test server
    print_status "Stopping test server..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
else
    print_error "Failed to start test server"
fi

# Clean up
rm -f smarapp-api-test
rm -f coverage.out

echo ""
print_success "Test suite completed!"
echo ""
echo "?? Test Summary:"
echo "  ? Unit Tests: Models, Middleware, Handlers"
echo "  ? Integration Tests: API endpoints"
echo "  ?? Coverage Report: coverage.html"
echo ""
echo "?? To run the application:"
echo "  go run cmd/server/main.go"
echo ""
echo "?? To view API documentation:"
echo "  Open http://localhost:8080/docs/index.html after starting the server"
