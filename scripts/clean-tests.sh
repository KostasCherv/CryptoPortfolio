#!/bin/bash

# Clean tests script for Simple API
# Removes test artifacts and temporary files

set -e

echo "🧹 Cleaning test artifacts..."
echo "============================="

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

# Files and directories to clean
FILES_TO_CLEAN=(
    "coverage.out"
    "coverage.html"
    "*.test"
    "*.prof"
    "*.trace"
)

DIRS_TO_CLEAN=(
    "coverage"
    "tmp"
    "testdata"
)

print_status "Cleaning coverage files..."
for file in "${FILES_TO_CLEAN[@]}"; do
    if ls $file 1> /dev/null 2>&1; then
        rm -f $file
        print_status "Removed: $file"
    fi
done

print_status "Cleaning test directories..."
for dir in "${DIRS_TO_CLEAN[@]}"; do
    if [ -d "$dir" ]; then
        rm -rf "$dir"
        print_status "Removed directory: $dir"
    fi
done

print_status "Cleaning Go cache..."
go clean -cache -testcache -modcache

print_status "Cleaning build artifacts..."
go clean -buildcache

print_success "Test cleanup completed!" 