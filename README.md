# CryptoPortfolio

A high-performance REST API built with Go, featuring wallet watchlist management, Web3 integration, and background balance fetching for crypto portfolio tracking.

## Features

- **Wallet Watchlist Management** - Add/remove wallets and tokens to track
- **Web3 Integration** - Real-time balance fetching from Ethereum blockchain
- **Background Processing** - Automated balance updates with configurable intervals
- **Redis Caching** - High-performance caching for API responses
- **JWT Authentication** - Secure user authentication
- **PostgreSQL Database** - Reliable data persistence
- **Swagger Documentation** - Interactive API documentation
- **Docker Support** - Easy deployment with Docker Compose

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL
- Redis

### 1. Clone and Setup
```bash
git clone <repository>
cd cryptoportfolio
cp docs/env.example .env
# Edit .env with your configuration
```

### 2. Start Services
```bash
# Using Docker Compose (recommended)
docker-compose up -d

# Or start manually
go run cmd/server/main.go
```

### 3. Access the API
- **API**: http://localhost:8080
- **Swagger Docs**: http://localhost:8080/swagger/index.html

## Environment Variables

Copy `docs/env.example` to `.env` and configure:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=cryptoportfolio

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key

# Web3 Settings
WEB3_RATE_LIMIT=5        # Requests per second
WEB3_MAX_WORKERS=3       # Concurrent workers
WEB3_FETCH_INTERVAL=5    # Balance fetch interval (seconds)
```

## API Endpoints

### Public Endpoints
- `GET /health` - Health check

### Authentication
- `POST /api/v1/auth/register` - Create account
- `POST /api/v1/auth/login` - Login

### User Management (Protected)
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update current user profile

### Watchlist Management (Protected)

#### Wallet Management
- `POST /api/v1/watchlist/wallets` - Add wallet
- `GET /api/v1/watchlist/wallets` - List wallets
- `DELETE /api/v1/watchlist/wallets/{id}` - Remove wallet

#### Token Management
- `POST /api/v1/watchlist/tokens` - Add token
- `GET /api/v1/watchlist/tokens` - List tokens
- `DELETE /api/v1/watchlist/tokens/{id}` - Remove token

#### Balance Management
- `GET /api/v1/watchlist/balances` - Get current balances
- `POST /api/v1/watchlist/balances/refresh` - Force refresh balances
- `GET /api/v1/watchlist/wallets/{wallet_id}/tokens/{token_id}/history` - Get balance history for specific wallet/token

## Background Processing

The API automatically fetches wallet balances in the background:

- **Configurable intervals** via `WEB3_FETCH_INTERVAL`
- **Rate limiting** to avoid API limits
- **Exponential backoff** for failed requests
- **Concurrent processing** with worker pools

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Stress test the API
./stress-test.sh
```

## Project Structure

```
cryptoportfolio/
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/             # HTTP handlers, routes, middleware
│   ├── config/          # Configuration management
│   ├── database/        # Database connection
│   ├── models/          # Data models
│   └── services/        # Business logic (Web3, watchlist, etc.)
├── pkg/                 # Shared packages
├── docs/                # Documentation and Swagger
└── scripts/             # Utility scripts
```

## Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Development

```bash
# Run server
go run cmd/server/main.go

# Generate Swagger docs
./scripts/generate-docs.sh

# Run tests
./scripts/test.sh
```
