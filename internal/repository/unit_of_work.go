package repository

import (
	"context"

	"gorm.io/gorm"
)

// UnitOfWork manages multiple repositories and transactions
type UnitOfWork interface {
	UserRepository() UserRepository
	WithTransaction(ctx context.Context, fn func(UnitOfWork) error) error
	Begin(ctx context.Context) (UnitOfWork, error)
	Commit() error
	Rollback() error
}

// unitOfWork implements the UnitOfWork interface
type unitOfWork struct {
	db            *gorm.DB
	userRepo      UserRepository
	transaction   *gorm.DB
	isTransaction bool
}

// NewUnitOfWork creates a new UnitOfWork instance
func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &unitOfWork{
		db:       db,
		userRepo: NewUserRepository(db),
	}
}

// UserRepository returns the user repository
func (uow *unitOfWork) UserRepository() UserRepository {
	if uow.isTransaction && uow.transaction != nil {
		return NewUserRepository(uow.transaction)
	}
	return uow.userRepo
}

// WithTransaction executes a function within a transaction
func (uow *unitOfWork) WithTransaction(ctx context.Context, fn func(UnitOfWork) error) error {
	return uow.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUnitOfWork := &unitOfWork{
			db:            tx,
			userRepo:      NewUserRepository(tx),
			transaction:   tx,
			isTransaction: true,
		}
		return fn(txUnitOfWork)
	})
}

// Begin starts a new transaction
func (uow *unitOfWork) Begin(ctx context.Context) (UnitOfWork, error) {
	tx := uow.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	
	return &unitOfWork{
		db:            tx,
		userRepo:      NewUserRepository(tx),
		transaction:   tx,
		isTransaction: true,
	}, nil
}

// Commit commits the current transaction
func (uow *unitOfWork) Commit() error {
	if uow.transaction != nil {
		return uow.transaction.Commit().Error
	}
	return nil
}

// Rollback rolls back the current transaction
func (uow *unitOfWork) Rollback() error {
	if uow.transaction != nil {
		return uow.transaction.Rollback().Error
	}
	return nil
}

// RepositoryManager implementation
type repositoryManager struct {
	unitOfWork UnitOfWork
}

// NewRepositoryManager creates a new RepositoryManager
func NewRepositoryManager(db *gorm.DB) RepositoryManager {
	return &repositoryManager{
		unitOfWork: NewUnitOfWork(db),
	}
}

// UserRepository returns the user repository
func (rm *repositoryManager) UserRepository() UserRepository {
	return rm.unitOfWork.UserRepository()
}
