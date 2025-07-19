#!/bin/bash

# Test script for Simple API
# Runs all tests with coverage and proper formatting

set -e

echo "🧪 Running tests for Simple API..."
echo "=================================="

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

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root."
    exit 1
fi

# Clean up any existing coverage files
rm -f coverage.out coverage.html

print_status "Running all tests..."
go test -v ./...

print_status "Running tests with coverage..."
go test -coverprofile=coverage.out ./...

# Check if tests passed
if [ $? -eq 0 ]; then
    print_success "All tests passed!"
else
    print_error "Some tests failed!"
    exit 1
fi

# Generate coverage report
print_status "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Display coverage summary
print_status "Coverage summary:"
go tool cover -func=coverage.out | tail -1

# Check coverage threshold (80%)
COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE >= 80" | bc -l) )); then
    print_success "Coverage is above 80% threshold ($COVERAGE%)"
else
    print_warning "Coverage is below 80% threshold ($COVERAGE%)"
fi

print_status "Coverage report generated: coverage.html"
print_success "Testing completed successfully!" 