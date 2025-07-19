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
- **Comprehensive API documentation with Swagger/OpenAPI**

## Project Structure

```
simple_api/
├── cmd/server/main.go          # Application entry point
├── docs/                       # Generated Swagger documentation
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
├── scripts/                    # Utility scripts
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
   go get github.com/swaggo/swag
   go get github.com/swaggo/gin-swagger
   go get github.com/swaggo/files
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

## API Documentation

This API includes comprehensive OpenAPI/Swagger documentation that provides:

- **Interactive API Explorer**: Test endpoints directly from the browser
- **Request/Response Examples**: See exactly what data to send and what you'll receive
- **Authentication Details**: Learn how to use JWT tokens
- **Error Codes**: Understand all possible error responses
- **Data Models**: View the structure of all request and response objects

### Accessing the Documentation

1. **Start the server**:
   ```bash
   go run cmd/server/main.go
   ```

2. **Open your browser** and navigate to:
   ```
   http://localhost:8080/swagger/index.html
   ```

3. **Explore the API**:
   - Browse all available endpoints organized by tags
   - Test endpoints directly with the "Try it out" feature
   - View request/response schemas and examples
   - Authenticate using the "Authorize" button

### API Endpoints Overview

#### Authentication
- `POST /api/v1/auth/register` - Create a new user account
- `POST /api/v1/auth/login` - Authenticate and receive JWT token

#### User Management
- `GET /api/v1/users/me` - Get current user profile (requires authentication)
- `PUT /api/v1/users/me` - Update current user profile (requires authentication)

#### Health Check
- `GET /health` - Check server status

### Authentication

Protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

To get a token:
1. Register a new user or login with existing credentials
2. Copy the token from the response
3. Click the "Authorize" button in Swagger UI
4. Enter: `Bearer <your-token>`
5. Now you can test protected endpoints

### Regenerating Documentation

If you modify the API endpoints or add new ones, regenerate the documentation:

```bash
./scripts/generate-docs.sh
```

Or manually:
```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs
```

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