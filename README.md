# Simple API

A high-performance REST API built with Go, featuring clean architecture and proper separation of concerns.

## Features

- High-performance HTTP handling with Gin
- Clean architecture with internal/external library separation
- Database integration with GORM and PostgreSQL
- JWT-based authentication
- Environment-based configuration
- Structured logging with Zap
- Graceful shutdown handling

## Project Structure

```
simple_api/
├── cmd/server/main.go          # Application entry point
├── internal/
│   ├── api/                    # HTTP handlers, middleware, routes
│   ├── config/                 # Configuration management
│   ├── database/               # Database connection and models
│   ├── models/                 # Data models
│   ├── services/               # Business logic layer
│   └── utils/                  # Internal utilities
├── pkg/
│   ├── logger/                 # Logging package
│   └── validator/              # Validation utilities
├── go.mod                      # Go module file
├── .env                        # Environment variables
└── README.md                   # This file
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher

## Installation

1. **Initialize Go module**
   ```bash
   go mod init simple_api
   ```

2. **Install dependencies**
   ```bash
   go get github.com/gin-gonic/gin
   go get gorm.io/gorm
   go get gorm.io/driver/postgres
   go get github.com/spf13/viper
   go get go.uber.org/zap
   go get github.com/golang-jwt/jwt/v5
   go get github.com/go-playground/validator/v10
   go get github.com/stretchr/testify
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

4. **Set up PostgreSQL database**
   ```sql
   CREATE DATABASE simple_api;
   CREATE USER simple_api_user WITH PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE simple_api TO simple_api_user;
   ```

## Running the Application

### Development
```bash
go run cmd/server/main.go
```

### Production
```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

The server starts on `http://localhost:8080` by default.

## API Endpoints

- `GET /health` - Health check
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/users/me` - Get current user (protected)
- `PUT /api/v1/users/me` - Update current user (protected)

## Configuration

Environment variables in `.env`:

```env
ENVIRONMENT=development
SERVER_PORT=8080

DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=password
DATABASE_DB_NAME=simple_api
DATABASE_SSL_MODE=disable

JWT_SECRET=your-super-secret-jwt-key-here
```

## Testing

```bash
go test ./...
go test -cover ./...
``` 