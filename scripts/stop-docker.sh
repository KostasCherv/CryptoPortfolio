#!/bin/bash

# Stop Simple API Docker Compose services

set -e

echo "🛑 Stopping Simple API Docker Compose services..."
echo ""

# Stop services
docker-compose down

echo "✅ Services stopped"
echo ""
echo "📝 To start again, run: ./scripts/run-docker.sh" 