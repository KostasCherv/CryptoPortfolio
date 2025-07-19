#!/bin/bash

# Generate Swagger documentation
echo "Generating Swagger documentation..."

# Remove existing docs
rm -rf docs/

# Generate new docs
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs

echo "Swagger documentation generated successfully!"
echo "You can now access the API documentation at: http://localhost:8080/swagger/index.html" 