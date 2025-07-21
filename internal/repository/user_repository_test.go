package repository

import (
	"context"
	"testing"

	"simple_api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Auto migrate models
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)
	
	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
	}

	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user first
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Find the user
	found, err := repo.FindByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Name, found.Name)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user first
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Find the user by email
	found, err := repo.FindByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Name, found.Name)
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user first
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Check if user exists
	exists, err := repo.ExistsByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existent user
	exists, err = repo.ExistsByEmail(ctx, "nonexistent@example.com")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create multiple users
	users := []*models.User{
		{Email: "user1@example.com", Password: "pass1", Name: "User 1"},
		{Email: "user2@example.com", Password: "pass2", Name: "User 2"},
		{Email: "user3@example.com", Password: "pass3", Name: "User 3"},
	}

	for _, user := range users {
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	// List users with pagination
	opts := &QueryOptions{
		Pagination: &Pagination{
			Limit:  2,
			Offset: 0,
		},
		OrderBy:  "created_at",
		OrderDir: "desc",
	}

	result, err := repo.List(ctx, opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Len(t, result.Data, 2)
	assert.True(t, result.HasNext)
	assert.False(t, result.HasPrev)
}

func TestUserRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create users with different names
	users := []*models.User{
		{Email: "john@example.com", Password: "pass1", Name: "John Doe"},
		{Email: "jane@example.com", Password: "pass2", Name: "Jane Smith"},
		{Email: "bob@example.com", Password: "pass3", Name: "Bob Johnson"},
	}

	for _, user := range users {
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	// Search for users with "John" in name
	opts := &QueryOptions{
		Pagination: &Pagination{
			Limit:  10,
			Offset: 0,
		},
	}

	result, err := repo.Search(ctx, "John", opts)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Total) // John Doe and Bob Johnson
	assert.Len(t, result.Data, 2)
}
