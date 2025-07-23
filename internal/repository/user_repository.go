package repository

import (
	"context"
	"errors"

	"cryptoportfolio/internal/models"

	"gorm.io/gorm"
)

// UserRepository defines the contract for user data access operations
type UserRepository interface {
	BaseRepository[models.User]
	
	// User-specific operations
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	List(ctx context.Context, opts *QueryOptions) (*PaginatedResult[models.User], error)
	Count(ctx context.Context) (int64, error)
	FindByIDs(ctx context.Context, ids []uint) ([]*models.User, error)
	Search(ctx context.Context, query string, opts *QueryOptions) (*PaginatedResult[models.User], error)
}

// userRepository implements the UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicateKey
		}
		return ErrDatabaseError
	}
	return nil
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrDatabaseError
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, ErrDatabaseError
	}
	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return ErrDatabaseError
	}
	return nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		return ErrDatabaseError
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Exists checks if a user exists with the given ID
func (r *userRepository) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, ErrDatabaseError
	}
	return count > 0, nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, ErrDatabaseError
	}
	return count > 0, nil
}

// List retrieves a paginated list of users
func (r *userRepository) List(ctx context.Context, opts *QueryOptions) (*PaginatedResult[models.User], error) {
	var users []*models.User
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.User{})
	
	// Apply filters
	if opts != nil && opts.Filters != nil {
		for key, value := range opts.Filters {
			query = query.Where(key+" = ?", value)
		}
	}
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, ErrDatabaseError
	}
	
	// Apply pagination
	if opts != nil && opts.Pagination != nil {
		query = query.Limit(opts.Pagination.Limit).Offset(opts.Pagination.Offset)
	}
	
	// Apply ordering
	if opts != nil && opts.OrderBy != "" {
		orderDir := "asc"
		if opts.OrderDir == "desc" {
			orderDir = "desc"
		}
		query = query.Order(opts.OrderBy + " " + orderDir)
	}
	
	// Execute query
	if err := query.Find(&users).Error; err != nil {
		return nil, ErrDatabaseError
	}
	
	// Build pagination result
	result := &PaginatedResult[models.User]{
		Data:   users,
		Total:  total,
		Limit:  opts.Pagination.Limit,
		Offset: opts.Pagination.Offset,
	}
	
	// Calculate pagination metadata
	if opts != nil && opts.Pagination != nil {
		result.HasNext = result.Offset+result.Limit < int(result.Total)
		result.HasPrev = result.Offset > 0
	}
	
	return result, nil
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, ErrDatabaseError
	}
	return count, nil
}

// FindByIDs finds users by their IDs
func (r *userRepository) FindByIDs(ctx context.Context, ids []uint) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, ErrDatabaseError
	}
	return users, nil
}

// Search searches users by name or email
func (r *userRepository) Search(ctx context.Context, query string, opts *QueryOptions) (*PaginatedResult[models.User], error) {
	var users []*models.User
	var total int64
	
	searchQuery := r.db.WithContext(ctx).Model(&models.User{}).
		Where("name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%")
	
	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, ErrDatabaseError
	}
	
	// Apply pagination
	if opts != nil && opts.Pagination != nil {
		searchQuery = searchQuery.Limit(opts.Pagination.Limit).Offset(opts.Pagination.Offset)
	}
	
	// Execute query
	if err := searchQuery.Find(&users).Error; err != nil {
		return nil, ErrDatabaseError
	}
	
	// Build pagination result
	result := &PaginatedResult[models.User]{
		Data:   users,
		Total:  total,
		Limit:  opts.Pagination.Limit,
		Offset: opts.Pagination.Offset,
	}
	
	// Calculate pagination metadata
	if opts != nil && opts.Pagination != nil {
		result.HasNext = result.Offset+result.Limit < int(result.Total)
		result.HasPrev = result.Offset > 0
	}
	
	return result, nil
}

// isDuplicateKeyError checks if the error is a duplicate key violation
func isDuplicateKeyError(err error) bool {
	// This is a simplified check - in production you might want to check specific error codes
	// depending on your database driver
	return err != nil && (err.Error() == "UNIQUE constraint failed: users.email" ||
		err.Error() == "duplicate key value violates unique constraint")
}
