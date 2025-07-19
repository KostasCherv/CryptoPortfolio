#!/bin/bash

# Run the Go API server

set -e

echo "Starting the API server..."

go run cmd/server/main.go
