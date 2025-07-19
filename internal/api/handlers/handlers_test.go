package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"simple_api/internal/config"
	"simple_api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a test database using SQLite
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	
	// Auto migrate the User model
	err = db.AutoMigrate(&models.User{})
	assert.NoError(t, err)
	
	return db
}

// setupTestHandler creates a handler with test dependencies
func setupTestHandler(t *testing.T) *Handler {
	db := setupTestDB(t)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-testing-only",
		},
	}
	return NewHandler(db, cfg)
}

// setupTestRouter creates a test router with the handler
func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/health", handler.HealthCheck)
	router.POST("/auth/register", handler.Register())
	router.POST("/auth/login", handler.Login())
	router.GET("/users/me", handler.GetCurrentUser())
	router.PUT("/users/me", handler.UpdateUser())
	
	return router
}

// generateTestJWT creates a JWT token for testing
func generateTestJWT(userID uint, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	return token.SignedString([]byte(secret))
}

// createTestUser creates a user with hashed password for testing
func createTestUser(t *testing.T, db *gorm.DB, email, password, name string) *models.User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	
	user := models.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
	}
	
	err = db.Create(&user).Error
	assert.NoError(t, err)
	
	return &user
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "Password123",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1",
			wantErr:  true,
		},
		{
			name:     "no uppercase",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "no lowercase",
			password: "PASSWORD123",
			wantErr:  true,
		},
		{
			name:     "no number",
			password: "Password",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.wantErr {
				assert.NotEmpty(t, err)
			} else {
				assert.Empty(t, err)
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	handler := setupTestHandler(t)
	router := setupTestRouter(handler)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "Password123",
			Name:     "Test User",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, "User registered successfully", response["message"])
		assert.NotEmpty(t, response["token"])
		assert.NotEmpty(t, response["user"])
	})

	t.Run("invalid password", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "test2@example.com",
			Password: "weak",
			Name:     "Test User",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["error"])
	})

	t.Run("duplicate email", func(t *testing.T) {
		// First registration
		reqBody := RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "Password123",
			Name:     "Test User",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		// Second registration with same email
		req2 := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
		req2.Header.Set("Content-Type", "application/json")
		
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		
		assert.Equal(t, http.StatusConflict, w2.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w2.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User with this email already exists", response["error"])
	})
}

func TestLoginHandler(t *testing.T) {
	handler := setupTestHandler(t)
	router := setupTestRouter(handler)

	t.Run("successful login", func(t *testing.T) {
		// Create a test user
		user := createTestUser(t, handler.DB, "login@example.com", "Password123", "Login User")
		
		reqBody := LoginRequest{
			Email:    "login@example.com",
			Password: "Password123",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, "Login successful", response["message"])
		assert.NotEmpty(t, response["token"])
		assert.NotEmpty(t, response["user"])
		
		// Verify user data in response
		userData := response["user"].(map[string]interface{})
		assert.Equal(t, float64(user.ID), userData["id"])
		assert.Equal(t, user.Email, userData["email"])
		assert.Equal(t, user.Name, userData["name"])
	})

	t.Run("user not found", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "Password123",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid credentials", response["error"])
	})

	t.Run("wrong password", func(t *testing.T) {
		// Create a test user
		createTestUser(t, handler.DB, "wrongpass@example.com", "Password123", "Wrong Pass User")
		
		reqBody := LoginRequest{
			Email:    "wrongpass@example.com",
			Password: "WrongPassword",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid credentials", response["error"])
	})

	t.Run("invalid request data", func(t *testing.T) {
		reqBody := map[string]string{
			"email": "invalid-email",
			// missing password
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request data", response["error"])
	})

	t.Run("case insensitive email", func(t *testing.T) {
		// Create user with lowercase email
		createTestUser(t, handler.DB, "case@example.com", "Password123", "Case User")
		
		// Try to login with uppercase email
		reqBody := LoginRequest{
			Email:    "CASE@EXAMPLE.COM",
			Password: "Password123",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Login successful", response["message"])
	})
}

func TestGetCurrentUser(t *testing.T) {
	handler := setupTestHandler(t)

	t.Run("successful get current user", func(t *testing.T) {
		// First create a user
		user := models.User{
			Email:    "test@example.com",
			Password: "hashedpassword",
			Name:     "Test User",
		}
		err := handler.DB.Create(&user).Error
		assert.NoError(t, err)

		// Generate JWT token
		token, err := generateTestJWT(user.ID, handler.Config.JWT.Secret)
		assert.NoError(t, err)

		// Create a custom context with user_id
		req := httptest.NewRequest("GET", "/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Create a custom gin context with user_id set
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		// Call the handler directly
		handler.GetCurrentUser()(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		userData := response["user"].(map[string]interface{})
		assert.Equal(t, float64(user.ID), userData["id"])
		assert.Equal(t, user.Email, userData["email"])
		assert.Equal(t, user.Name, userData["name"])
	})

	t.Run("user not found", func(t *testing.T) {
		// Generate JWT token for non-existent user
		token, err := generateTestJWT(999, handler.Config.JWT.Secret)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", uint(999))

		handler.GetCurrentUser()(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not found", response["error"])
	})
}

func TestUpdateUser(t *testing.T) {
	handler := setupTestHandler(t)

	t.Run("successful update user", func(t *testing.T) {
		// First create a user
		user := models.User{
			Email:    "update@example.com",
			Password: "hashedpassword",
			Name:     "Original Name",
		}
		err := handler.DB.Create(&user).Error
		assert.NoError(t, err)

		// Update request
		updateReq := UpdateUserRequest{
			Name: "Updated Name",
		}
		jsonBody, _ := json.Marshal(updateReq)

		req := httptest.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		handler.UpdateUser()(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "User updated successfully", response["message"])
		
		userData := response["user"].(map[string]interface{})
		assert.Equal(t, "Updated Name", userData["name"])
		assert.Equal(t, user.Email, userData["email"])
	})

	t.Run("invalid request data", func(t *testing.T) {
		// Create a user
		user := models.User{
			Email:    "invalid@example.com",
			Password: "hashedpassword",
			Name:     "Test User",
		}
		err := handler.DB.Create(&user).Error
		assert.NoError(t, err)

		// Invalid request (empty name)
		invalidReq := map[string]string{"name": ""}
		jsonBody, _ := json.Marshal(invalidReq)

		req := httptest.NewRequest("PUT", "/users/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", user.ID)

		handler.UpdateUser()(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request data", response["error"])
	})
}

func TestHealthCheck(t *testing.T) {
	handler := setupTestHandler(t)
	router := setupTestRouter(handler)
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Server is running", response["message"])
	assert.NotEmpty(t, response["time"])
} 