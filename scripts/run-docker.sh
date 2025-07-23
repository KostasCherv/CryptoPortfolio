#!/bin/bash

# Run Simple API with Docker Compose (PostgreSQL + Redis)

set -e

echo "🚀 Starting Simple API with Docker Compose..."
echo "📦 Services: PostgreSQL + Redis"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Copy environment file if it doesn't exist
if [ ! -f .env ]; then
    echo "📋 Creating .env file from docker.env..."
    cp docker.env .env
    echo "✅ .env file created"
fi

# Start services
echo "🔧 Starting services..."
docker-compose up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
echo "   - PostgreSQL..."
until docker-compose exec -T db pg_isready -U postgres > /dev/null 2>&1; do
    sleep 2
done
echo "   ✅ PostgreSQL is ready"

echo "   - Redis..."
until docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; do
    sleep 2
done
echo "   ✅ Redis is ready"

echo ""
echo "🎉 All services are running!"
echo ""
echo "📊 Service Status:"
docker-compose ps
echo ""
echo "🌐 Application will be available at: http://localhost:8080"
echo "📚 Swagger docs: http://localhost:8080/swagger/index.html"
echo "🗄️  PostgreSQL: localhost:5432"
echo "🔴 Redis: localhost:6379"
echo ""
echo "📝 Useful commands:"
echo "   docker-compose logs -f    # View logs"
echo "   docker-compose down       # Stop services"
echo "   docker-compose restart    # Restart services" 