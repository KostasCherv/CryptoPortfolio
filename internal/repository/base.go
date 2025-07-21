package repository

import (
	"context"
	"errors"
)

// Common repository errors
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrDuplicateKey   = errors.New("duplicate key violation")
	ErrDatabaseError  = errors.New("database error")
	ErrInvalidInput   = errors.New("invalid input")
)

// BaseRepository defines common operations for all repositories
type BaseRepository[T any] interface {
	Create(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id uint) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
	Exists(ctx context.Context, id uint) (bool, error)
}

// Pagination represents pagination parameters
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PaginatedResult represents a paginated result
type PaginatedResult[T any] struct {
	Data       []*T `json:"data"`
	Total      int64 `json:"total"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// QueryOptions represents query options for repositories
type QueryOptions struct {
	Pagination *Pagination
	OrderBy    string
	OrderDir   string // "asc" or "desc"
	Filters    map[string]interface{}
}

// RepositoryManager manages multiple repositories
type RepositoryManager interface {
	UserRepository() UserRepository
	// Add more repositories here as needed
	// ProductRepository() ProductRepository
	// OrderRepository() OrderRepository
}

// TransactionManager handles database transactions
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Begin(ctx context.Context) (Transaction, error)
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
	UserRepository() UserRepository
}
