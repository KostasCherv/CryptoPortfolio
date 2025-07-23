#!/bin/bash

# Test coverage script for CryptoPortfolio
# Generates detailed coverage reports

set -e

echo "📊 Generating test coverage report..."
echo "===================================="

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

# Create coverage directory
mkdir -p coverage

# Clean up existing coverage files
rm -f coverage.out coverage.html coverage/coverage.txt

print_status "Running tests with coverage..."
go test -coverprofile=coverage.out ./...

# Generate detailed coverage report
print_status "Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Generate function-level coverage report
print_status "Generating function-level coverage report..."
go tool cover -func=coverage.out > coverage/coverage.txt

# Display coverage summary
print_status "Coverage summary:"
echo "=================="
go tool cover -func=coverage.out | tail -1

# Display coverage by package
print_status "Coverage by package:"
echo "======================"
go test -cover ./... | grep -E "(PASS|FAIL|coverage:)" | while read line; do
    if [[ $line == *"coverage:"* ]]; then
        echo "$line"
    fi
done

# Check coverage threshold
COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE >= 80" | bc -l) )); then
    print_success "Coverage is above 80% threshold ($COVERAGE%)"
else
    print_warning "Coverage is below 80% threshold ($COVERAGE%)"
fi

print_status "Coverage files generated:"
echo "- coverage.html (HTML report)"
echo "- coverage/coverage.txt (Function-level report)"
echo "- coverage.out (Raw coverage data)"

print_success "Coverage report generation completed!" 