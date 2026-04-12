package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rss-reader/internal/handlers"
	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"rss-reader/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDBHandlers(t *testing.T) *gorm.DB {
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

func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	tagRepo := repository.NewTagRepository(db)

	// Services
	authService := services.NewAuthService(userRepo, "test-secret")
	rssService := services.NewRSSService(feedRepo, articleRepo, nil)

	// Auth routes
	router.POST("/api/auth/register", handlers.Register(authService))
	router.POST("/api/auth/login", handlers.Login(authService))

	// Feed routes (protected)
	router.GET("/api/feeds", handlers.AuthMiddleware("test-secret"), handlers.GetFeeds(feedRepo))
	router.POST("/api/feeds", handlers.AuthMiddleware("test-secret"), handlers.CreateFeed(feedRepo, rssService))
	router.PUT("/api/feeds/:id", handlers.AuthMiddleware("test-secret"), handlers.UpdateFeed(feedRepo))
	router.DELETE("/api/feeds/:id", handlers.AuthMiddleware("test-secret"), handlers.DeleteFeed(feedRepo))

	// Article routes (protected)
	router.GET("/api/articles", handlers.AuthMiddleware("test-secret"), handlers.GetArticles(articleRepo))
	router.GET("/api/articles/search", handlers.AuthMiddleware("test-secret"), handlers.SearchArticles(articleRepo))
	router.PUT("/api/articles/:id/read", handlers.AuthMiddleware("test-secret"), handlers.MarkArticleRead(articleRepo))
	router.POST("/api/articles/mark-read", handlers.AuthMiddleware("test-secret"), handlers.MarkAllRead(articleRepo))
	router.GET("/api/articles/unread-count", handlers.AuthMiddleware("test-secret"), handlers.GetUnreadCount(articleRepo))

	// Tag routes (protected)
	router.GET("/api/tags", handlers.AuthMiddleware("test-secret"), handlers.GetTags(tagRepo))
	router.POST("/api/tags", handlers.AuthMiddleware("test-secret"), handlers.CreateTag(tagRepo))
	router.DELETE("/api/tags/:id", handlers.AuthMiddleware("test-secret"), handlers.DeleteTag(tagRepo))

	// Article Tag routes (protected)
	router.POST("/api/articles/tags", handlers.AuthMiddleware("test-secret"), handlers.AddArticleTag(articleRepo))
	router.DELETE("/api/articles/tags", handlers.AuthMiddleware("test-secret"), handlers.RemoveArticleTag(articleRepo))

	return router
}

// Auth Handlers Tests

func TestAuthHandler_Register_Success(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "id")
	assert.Contains(t, response, "email")
	assert.Equal(t, "test@example.com", response["email"])
}

func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	reqBody := map[string]string{
		"email":    "invalid-email",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
}

func TestAuthHandler_Register_ShortPassword(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register first user
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Try to register again with same email
	req, _ = http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"], "already")
}

func TestAuthHandler_Login_Success(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register first
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Login
	loginReqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReqBody)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "token")
	assert.Contains(t, response, "user")
	assert.NotEmpty(t, response["token"])

	user := response["user"].(map[string]interface{})
	assert.Contains(t, user, "id")
	assert.Contains(t, user, "email")
	assert.Equal(t, "test@example.com", user["email"])
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	reqBody := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
}

func TestAuthHandler_AuthMiddleware_MissingHeader(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	req, _ := http.NewRequest("GET", "/api/feeds", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
}

func TestAuthHandler_AuthMiddleware_InvalidToken(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	req, _ := http.NewRequest("GET", "/api/feeds", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
}

// Feed Handlers Tests

func TestFeedHandler_GetFeeds_Empty(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register and login to get token
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)

	// Login
	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	token := loginResponse["token"].(string)

	// Get feeds
	req, _ = http.NewRequest("GET", "/api/feeds", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Empty(t, response)
}

func TestFeedHandler_CreateFeed_Success(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register and login
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	token := loginResponse["token"].(string)

	// Create feed
	feedReq := map[string]string{
		"url":      "https://example.com/feed.xml",
		"title":    "Test Feed",
		"category": "tech",
	}
	jsonBody, _ = json.Marshal(feedReq)

	req, _ = http.NewRequest("POST", "/api/feeds", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "id")
	assert.Equal(t, "https://example.com/feed.xml", response["url"])
	assert.Equal(t, "Test Feed", response["title"])
}

func TestFeedHandler_CreateFeed_MissingURL(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register and login
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	token := loginResponse["token"].(string)

	// Create feed without URL
	feedReq := map[string]string{
		"title":    "Test Feed",
		"category": "tech",
	}
	jsonBody, _ = json.Marshal(feedReq)

	req, _ = http.NewRequest("POST", "/api/feeds", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
}

// Article Handlers Tests

func TestArticleHandler_GetArticles_Empty(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register and login
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	token := loginResponse["token"].(string)

	// Get articles
	req, _ = http.NewRequest("GET", "/api/articles", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "articles")
	assert.Contains(t, response, "total")
	assert.Equal(t, float64(0), response["total"])

	articles := response["articles"].([]interface{})
	assert.Empty(t, articles)
}

func TestArticleHandler_SearchArticles_NoQuery(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register and login
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	token := loginResponse["token"].(string)

	// Search without query
	req, _ = http.NewRequest("GET", "/api/articles/search", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
}

// Tag Handlers Tests

func TestTagHandler_GetTags_Empty(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register and login
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	token := loginResponse["token"].(string)

	// Get tags
	req, _ = http.NewRequest("GET", "/api/tags", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Empty(t, response)
}

func TestTagHandler_CreateTag_Success(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register and login
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	token := loginResponse["token"].(string)

	// Create tag
	tagReq := map[string]string{
		"name": "AI",
	}
	jsonBody, _ = json.Marshal(tagReq)

	req, _ = http.NewRequest("POST", "/api/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "id")
	assert.Equal(t, "AI", response["name"])
}

func TestTagHandler_CreateTag_MissingName(t *testing.T) {
	db := setupTestDBHandlers(t)
	router := setupRouter(db)

	// Register and login
	registerReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	token := loginResponse["token"].(string)

	// Create tag without name
	tagReq := map[string]string{}
	jsonBody, _ = json.Marshal(tagReq)

	req, _ = http.NewRequest("POST", "/api/tags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "error")
}
