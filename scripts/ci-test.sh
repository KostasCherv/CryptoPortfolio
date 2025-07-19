#!/bin/bash

# CI test script for Simple API
# Runs all necessary checks for continuous integration

set -e

echo "🚀 Running CI tests..."
echo "======================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root."
    exit 1
fi

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Step 1: Format code
print_status "Step 1: Formatting code..."
go fmt ./...
print_success "Code formatting completed"

# Step 2: Vet code
print_status "Step 2: Vetting code..."
go vet ./...
print_success "Code vetting completed"

# Step 3: Run linter (if available)
if command_exists golangci-lint; then
    print_status "Step 3: Running linter..."
    golangci-lint run
    print_success "Linting completed"
else
    print_warning "golangci-lint not found, skipping linting"
fi

# Step 4: Run security scanner (if available)
if command_exists gosec; then
    print_status "Step 4: Running security scan..."
    gosec ./...
    print_success "Security scan completed"
else
    print_warning "gosec not found, skipping security scan"
fi

# Step 5: Run tests
print_status "Step 5: Running tests..."
go test -v ./...

# Step 6: Run tests with coverage
print_status "Step 6: Running tests with coverage..."
go test -coverprofile=coverage.out ./...

# Check test results
if [ $? -eq 0 ]; then
    print_success "All tests passed!"
else
    print_error "Some tests failed!"
    exit 1
fi

# Step 7: Check coverage threshold
print_status "Step 7: Checking coverage threshold..."
COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE >= 80" | bc -l) )); then
    print_success "Coverage is above 80% threshold ($COVERAGE%)"
else
    print_error "Coverage is below 80% threshold ($COVERAGE%)"
    exit 1
fi

# Step 8: Build the application
print_status "Step 8: Building application..."
go build -o bin/server cmd/server/main.go
print_success "Build completed"

# Step 9: Run integration tests (if they exist)
if [ -f "scripts/test-integration.sh" ]; then
    print_status "Step 9: Running integration tests..."
    ./scripts/test-integration.sh
    print_success "Integration tests completed"
else
    print_warning "Integration test script not found, skipping"
fi

print_success "All CI checks passed! 🎉" 