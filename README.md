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
- **Extensive test coverage with automated testing**

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

## Testing

This project includes comprehensive testing with high coverage and automated test scripts.

### Quick Test Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests for specific package
go test ./internal/api/handlers
go test ./internal/api/middleware
go test ./internal/config
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# Check coverage percentage
go test -cover ./... | grep -E "(PASS|FAIL|coverage:)"
```

### Testing Scripts

The project includes several testing scripts in the `scripts/` directory:

```bash
# Run all tests with coverage
./scripts/test.sh

# Run tests and generate coverage report
./scripts/test-coverage.sh

# Run tests in watch mode (requires fswatch)
./scripts/test-watch.sh

# Run integration tests
./scripts/test-integration.sh

# Clean test artifacts
./scripts/clean-tests.sh
```

### Test Database

Tests use SQLite in-memory database for fast execution:

```bash
# Run tests with specific database
DATABASE_DRIVER=sqlite go test ./...

# Run tests with PostgreSQL (requires running DB)
DATABASE_DRIVER=postgres go test ./...
```

### API Testing

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test user registration
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123","name":"Test User"}'

# Test user login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123"}'
```

### Performance Testing

```bash
# Run benchmark tests
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkRegister ./internal/api/handlers

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...
```

### Test Best Practices

1. **Test Structure**: Each test follows the Arrange-Act-Assert pattern
2. **Test Isolation**: Each test is independent and cleans up after itself
3. **Mock Usage**: External dependencies are mocked for unit tests
4. **Coverage Target**: Aim for >80% test coverage
5. **Test Naming**: Tests are named descriptively (e.g., `TestRegisterHandler_SuccessfulRegistration`)

### Continuous Integration

The project includes CI/CD configuration for automated testing:

```bash
# Run CI tests locally
./scripts/ci-test.sh

# Run linting
golangci-lint run

# Run security scanning
gosec ./...
```

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
