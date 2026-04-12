package repository_test

import (
	"rss-reader/internal/repository"
	"testing"
)

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	user, err := repo.Create("test@example.com", "password123")

	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should not be 0")
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.Password == "password123" {
		t.Error("Password should be hashed, not plain text")
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	// 创建第一个用户
	_, err := repo.Create("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// 尝试创建重复邮箱的用户
	_, err = repo.Create("test@example.com", "anotherpassword")
	if err == nil {
		t.Error("Expected error when creating user with duplicate email")
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	// 创建用户
	createdUser := createTestUser(db, "test@example.com")

	// 查找用户
	foundUser, err := repo.FindByEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to find user by email: %v", err)
	}

	if foundUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, foundUser.ID)
	}

	if foundUser.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", foundUser.Email)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	// 查找不存在的用户
	_, err := repo.FindByEmail("nonexistent@example.com")
	if err == nil {
		t.Error("Expected error when finding non-existent user")
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	// 创建用户
	createdUser := createTestUser(db, "test@example.com")

	// 查找用户
	foundUser, err := repo.FindByID(createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to find user by ID: %v", err)
	}

	if foundUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, foundUser.ID)
	}

	if foundUser.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", foundUser.Email)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	// 查找不存在的用户
	_, err := repo.FindByID(99999)
	if err == nil {
		t.Error("Expected error when finding non-existent user")
	}
}

func TestUserRepository_MultipleUsers(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	// 创建多个用户
	emails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for _, email := range emails {
		_, err := repo.Create(email, "password123")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// 验证所有用户都能找到
	for _, email := range emails {
		_, err := repo.FindByEmail(email)
		if err != nil {
			t.Errorf("Failed to find user with email %s: %v", email, err)
		}
	}
}
