#!/bin/bash

# Test watch script for CryptoPortfolio
# Runs tests automatically when files change

set -e

echo "👀 Starting test watcher..."
echo "==========================="

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

# Check if fswatch is installed
if ! command -v fswatch &> /dev/null; then
    print_error "fswatch is not installed. Please install it first:"
    echo "  macOS: brew install fswatch"
    echo "  Linux: sudo apt-get install fswatch"
    echo "  Or use: go install github.com/cespare/reflex@latest"
    exit 1
fi

print_status "Watching for file changes..."
print_status "Press Ctrl+C to stop"

# Function to run tests
run_tests() {
    echo ""
    echo "🔄 File changed, running tests..."
    echo "================================="
    
    # Run tests with coverage
    go test -cover ./...
    
    if [ $? -eq 0 ]; then
        print_success "Tests passed!"
    else
        print_error "Tests failed!"
    fi
    
    echo ""
    print_status "Watching for changes... (Press Ctrl+C to stop)"
}

# Watch for changes in Go files
fswatch -o \
    --exclude='.*' \
    --include='\.go$' \
    --include='\.mod$' \
    --include='\.sum$' \
    . | while read f; do
    run_tests
done 