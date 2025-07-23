package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"cryptoportfolio/internal/cache"
	"cryptoportfolio/internal/config"
	"cryptoportfolio/internal/models"
	"cryptoportfolio/internal/repository"
	"cryptoportfolio/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Common errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrTokenGeneration   = errors.New("failed to generate token")
)

// Request/Response types for the service layer
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthResponse struct {
	Message string       `json:"message"`
	Token   string       `json:"token"`
	User    UserResponse `json:"user"`
}

// UserService interface defines the contract for user-related business logic
type UserService interface {
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	GetUserByID(ctx context.Context, userID uint) (*UserResponse, error)
	UpdateUser(ctx context.Context, userID uint, req *UpdateUserRequest) (*UserResponse, error)
	ListUsers(ctx context.Context, opts *repository.QueryOptions) (*repository.PaginatedResult[UserResponse], error)
	SearchUsers(ctx context.Context, query string, opts *repository.QueryOptions) (*repository.PaginatedResult[UserResponse], error)
	ValidatePassword(password string) error
	GenerateJWT(userID uint) (string, error)
}

// userService implements the UserService interface
type userService struct {
	userRepo   repository.UserRepository
	userCache  cache.UserCacheProvider
	config     *config.Config
	logger     *logger.Logger
}

// NewUserService creates a new instance of UserService
func NewUserService(userRepo repository.UserRepository, userCache cache.UserCacheProvider, config *config.Config, logger *logger.Logger) UserService {
	return &userService{
		userRepo:  userRepo,
		userCache: userCache,
		config:    config,
		logger:    logger,
	}
}

// Register handles user registration business logic
func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Validate password strength
	if err := s.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if user already exists using repository
	exists, err := s.userRepo.ExistsByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		s.logger.Error("Failed to check if user exists", "error", err, "email", req.Email)
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, ErrInvalidPassword
	}

	// Create user using repository
	user := &models.User{
		Email:    strings.ToLower(req.Email),
		Password: string(hashedPassword),
		Name:     strings.TrimSpace(req.Name),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return nil, ErrUserAlreadyExists
		}
		s.logger.Error("Failed to create user", "error", err)
		return nil, err
	}

	// Generate JWT token
	token, err := s.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("User registered successfully", "user_id", user.ID, "email", user.Email)

	return &AuthResponse{
		Message: "User registered successfully",
		Token:   token,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// Login handles user authentication business logic
func (s *userService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Find user by email using repository
	user, err := s.userRepo.FindByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("Database error during login", "error", err)
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("User logged in successfully", "user_id", user.ID, "email", user.Email)

	return &AuthResponse{
		Message: "Login successful",
		Token:   token,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(ctx context.Context, userID uint) (*UserResponse, error) {
	// Try cache first
	cachedUser, err := s.userCache.GetUserByID(ctx, userID)
	if err == nil {
		s.logger.Debug("User found in cache", "user_id", userID)
		return &UserResponse{
			ID:        cachedUser.ID,
			Email:     cachedUser.Email,
			Name:      cachedUser.Name,
			CreatedAt: cachedUser.CreatedAt,
			UpdatedAt: cachedUser.UpdatedAt,
		}, nil
	}

	// Cache miss, get from database
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		s.logger.Error("Database error getting user", "error", err, "user_id", userID)
		return nil, err
	}

	// Store in cache for next time
	if err := s.userCache.SetUserByID(ctx, user); err != nil {
		s.logger.Warn("Failed to cache user", "error", err, "user_id", userID)
	}

	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateUser updates a user's profile
func (s *userService) UpdateUser(ctx context.Context, userID uint, req *UpdateUserRequest) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		s.logger.Error("Database error getting user for update", "error", err, "user_id", userID)
		return nil, err
	}

	// Update user
	user.Name = strings.TrimSpace(req.Name)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user", "error", err, "user_id", userID)
		return nil, err
	}

	// Invalidate cache
	if err := s.userCache.InvalidateUser(ctx, user.ID, user.Email); err != nil {
		s.logger.Warn("Failed to invalidate user cache", "error", err, "user_id", userID)
	}

	s.logger.Info("User updated successfully", "user_id", user.ID)

	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// ListUsers retrieves a paginated list of users
func (s *userService) ListUsers(ctx context.Context, opts *repository.QueryOptions) (*repository.PaginatedResult[UserResponse], error) {
	result, err := s.userRepo.List(ctx, opts)
	if err != nil {
		s.logger.Error("Failed to list users", "error", err)
		return nil, err
	}

	// Convert models to responses
	userResponses := make([]*UserResponse, len(result.Data))
	for i, user := range result.Data {
		userResponses[i] = &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	return &repository.PaginatedResult[UserResponse]{
		Data:    userResponses,
		Total:   result.Total,
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasNext: result.HasNext,
		HasPrev: result.HasPrev,
	}, nil
}

// SearchUsers searches users by name or email
func (s *userService) SearchUsers(ctx context.Context, query string, opts *repository.QueryOptions) (*repository.PaginatedResult[UserResponse], error) {
	result, err := s.userRepo.Search(ctx, query, opts)
	if err != nil {
		s.logger.Error("Failed to search users", "error", err, "query", query)
		return nil, err
	}

	// Convert models to responses
	userResponses := make([]*UserResponse, len(result.Data))
	for i, user := range result.Data {
		userResponses[i] = &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	return &repository.PaginatedResult[UserResponse]{
		Data:    userResponses,
		Total:   result.Total,
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasNext: result.HasNext,
		HasPrev: result.HasPrev,
	}, nil
}

// ValidatePassword validates password strength
func (s *userService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	
	// You can add more validation rules here
	// For now, keeping it simple to match the original implementation
	
	return nil
}

// GenerateJWT generates a JWT token for a user
func (s *userService) GenerateJWT(userID uint) (string, error) {
	if s.config.JWT.Secret == "" {
		s.logger.Error("JWT secret is empty", "user_id", userID)
		return "", ErrTokenGeneration
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":     time.Now().Unix(),
	})
	
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		s.logger.Error("Failed to generate JWT token", "error", err, "user_id", userID)
		return "", ErrTokenGeneration
	}
	
	return tokenString, nil
}
 