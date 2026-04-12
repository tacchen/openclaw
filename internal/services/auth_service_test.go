package services_test

import (
	"testing"

	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"rss-reader/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=rss_test port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up tables before test
	db.Exec("DROP TABLE IF EXISTS article_tags CASCADE")
	db.Exec("DROP TABLE IF EXISTS tags CASCADE")
	db.Exec("DROP TABLE IF EXISTS articles CASCADE")
	db.Exec("DROP TABLE IF EXISTS feeds CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")

	// Auto migrate tables
	if err := db.AutoMigrate(&models.User{}, &models.Feed{}, &models.Article{}, &models.Tag{}, &models.ArticleTag{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestAuthService_Register_Success(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	service := services.NewAuthService(userRepo, "test-secret")

	// Execute
	user, err := service.Register("test@example.com", "password123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEmpty(t, user.Password)
	assert.NotEqual(t, "password123", user.Password) // Password should be hashed
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	service := services.NewAuthService(userRepo, "test-secret")

	// Register first user
	_, err := service.Register("test@example.com", "password123")
	assert.NoError(t, err)

	// Try to register with same email
	user, err := service.Register("test@example.com", "different_password")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "already registered")
}

func TestAuthService_Login_Success(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	service := services.NewAuthService(userRepo, "test-secret")

	// Register a user first
	_, err := service.Register("test@example.com", "password123")
	assert.NoError(t, err)

	// Execute login
	token, user, err := service.Login("test@example.com", "password123")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	service := services.NewAuthService(userRepo, "test-secret")

	// Execute login with non-existent user
	token, user, err := service.Login("nonexistent@example.com", "password123")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	service := services.NewAuthService(userRepo, "test-secret")

	// Register a user first
	_, err := service.Register("test@example.com", "password123")
	assert.NoError(t, err)

	// Execute login with wrong password
	token, user, err := service.Login("test@example.com", "wrongpassword")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_Login_MultipleUsers(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	service := services.NewAuthService(userRepo, "test-secret")

	// Register multiple users
	user1, err := service.Register("user1@example.com", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, user1)

	user2, err := service.Register("user2@example.com", "password456")
	assert.NoError(t, err)
	assert.NotNil(t, user2)

	// Login as user1
	token1, loggedUser1, err := service.Login("user1@example.com", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)
	assert.Equal(t, user1.ID, loggedUser1.ID)

	// Login as user2
	token2, loggedUser2, err := service.Login("user2@example.com", "password456")
	assert.NoError(t, err)
	assert.NotEmpty(t, token2)
	assert.Equal(t, user2.ID, loggedUser2.ID)

	// Tokens should be different
	assert.NotEqual(t, token1, token2)
}

func TestAuthService_TokenClaims(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	service := services.NewAuthService(userRepo, "test-secret")

	// Register and login
	_, err := service.Register("test@example.com", "password123")
	assert.NoError(t, err)

	token, user, err := service.Login("test@example.com", "password123")
	assert.NoError(t, err)

	// Parse token and verify claims
	claims := &services.Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if parsedToken.Valid {
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.ID, claims.GetUserID())
	} else {
		t.Fatal("Invalid token")
	}
}
