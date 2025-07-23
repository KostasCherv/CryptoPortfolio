#!/bin/bash

# Stop CryptoPortfolio Docker Compose services

set -e

echo "🛑 Stopping CryptoPortfolio Docker Compose services..."
echo ""

# Stop services
docker-compose down

echo "✅ Services stopped"
echo ""
echo "📝 To start again, run: ./scripts/run-docker.sh" 