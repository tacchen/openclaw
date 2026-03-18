package handlers

import (
	"encoding/json"
	"net/http"
	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"rss-reader/internal/services"
	"strconv"

	"rss-reader/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Auth handlers
func Register(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := authService.Register(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":    user.ID,
			"email": user.Email,
		})
	}
}

func Login(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, user, err := authService.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":    user.ID,
				"email": user.Email,
			},
		})
	}
}

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// Feed handlers
func GetFeeds(feedRepo *repository.FeedRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		feeds, err := feedRepo.FindByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, feeds)
	}
}

func CreateFeed(feedRepo *repository.FeedRepository, rssService *services.RSSService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var req struct {
			URL      string `json:"url" binding:"required"`
			Title    string `json:"title"`
			Category string `json:"category"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if feed already exists for this user
		existing, _ := feedRepo.FindByURLAndUserID(req.URL, userID)
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "该订阅源已存在"})
			return
		}

		feed := &models.Feed{
			URL:      req.URL,
			Title:    req.Title,
			Category: req.Category,
			IconURL:  utils.GetFaviconURL(req.URL),
			UserID:   userID,
		}

		if err := feedRepo.Create(feed); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Fetch articles asynchronously (don't block the response)
		go rssService.FetchAndSaveArticles(feed)

		c.JSON(http.StatusCreated, feed)
	}
}

func UpdateFeed(feedRepo *repository.FeedRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
			return
		}

		feed, err := feedRepo.FindByID(uint(id))
		if err != nil || feed.UserID != userID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feed not found"})
			return
		}

		var req struct {
			URL      string `json:"url"`
			Title    string `json:"title"`
			Category string `json:"category"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Always update fields (allow empty strings to clear values)
		feed.URL = req.URL
		feed.Title = req.Title
		feed.Category = req.Category

		if err := feedRepo.Update(feed); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, feed)
	}
}

func DeleteFeed(feedRepo *repository.FeedRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
			return
		}

		feed, err := feedRepo.FindByID(uint(id))
		if err != nil || feed.UserID != userID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feed not found"})
			return
		}

		if err := feedRepo.Delete(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Feed deleted"})
	}
}

// Article handlers
func GetArticles(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		feedIDStr := c.Query("feed_id")
		tagIDStr := c.Query("tag_id")
		category := c.Query("category")
		categories := c.Query("categories") // 支持多选分类（逗号分隔）
		isReadStr := c.Query("is_read")
		hasSummaryStr := c.Query("has_summary")

		var feedID uint
		if feedIDStr != "" {
			id, err := strconv.ParseUint(feedIDStr, 10, 32)
			if err == nil {
				feedID = uint(id)
			}
		}

		var tagID uint
		if tagIDStr != "" {
			id, err := strconv.ParseUint(tagIDStr, 10, 32)
			if err == nil {
				tagID = uint(id)
			}
		}

		var isRead *bool
		if isReadStr != "" {
			val := isReadStr == "true"
			isRead = &val
		}

		var hasSummary *bool
		if hasSummaryStr != "" {
			val := hasSummaryStr == "true"
			hasSummary = &val
		}

		// 如果有 categories 参数，优先使用它
		filterCategory := category
		if categories != "" {
			filterCategory = categories // 传递逗号分隔的多个分类
		}
		
		articles, total, err := articleRepo.FindByUserID(userID, page, pageSize, feedID, tagID, filterCategory, isRead, hasSummary)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"articles": articles,
			"total":    total,
			"page":     page,
			"per_page": pageSize,
		})
	}
}

func SearchArticles(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		query := c.Query("q")

		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
			return
		}

		articles, err := articleRepo.SearchByTitle(userID, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"articles": articles,
			"total":    len(articles),
		})
	}
}

// MarkArticleRead 标记单篇文章已读
func MarkArticleRead(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		if err := articleRepo.MarkAsRead(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Article marked as read"})
	}
}

// MarkAllRead 批量标记已读
func MarkAllRead(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var req struct {
			FeedID     uint   `json:"feed_id"`
			Category   string `json:"category"`
			Categories string `json:"categories"` // 支持多选分类（逗号分隔）
		}
		c.ShouldBindJSON(&req)

		// 优先使用 categories 参数
		filterCategory := req.Category
		if req.Categories != "" {
			filterCategory = req.Categories
		}

		count, err := articleRepo.MarkAllAsRead(userID, req.FeedID, filterCategory)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Articles marked as read",
			"count":   count,
		})
	}
}

// GetUnreadCount 获取未读数量统计
func GetUnreadCount(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		total, byFeed, byCategory, err := articleRepo.GetUnreadCount(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total":       total,
			"by_feed":     byFeed,
			"by_category": byCategory,
		})
	}
}

// GenerateArticleSummary 生成文章的 AI 概览
func GenerateArticleSummary(articleRepo *repository.ArticleRepository, openaiService *services.OpenAIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		// 获取文章
		article, err := articleRepo.GetArticleByID(uint(id))
		if err != nil || article.UserID != userID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		// 检查是否已有摘要
		if article.Summary != "" {
			// 解析 key_points
			var keyPoints []string
			json.Unmarshal([]byte(article.KeyPoints), &keyPoints)
			c.JSON(http.StatusOK, gin.H{
				"summary":    article.Summary,
				"key_points": keyPoints,
				"cached":     true,
			})
			return
		}

		// 生成摘要
		req := services.SummaryRequest{
			Title:       article.Title,
			Description: article.Description,
			Content:     article.Content,
		}
		result, err := openaiService.GenerateSummary(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate summary: " + err.Error()})
			return
		}

		// 保存到数据库
		keyPointsJSON, _ := json.Marshal(result.KeyPoints)
		if err := articleRepo.UpdateArticleSummary(uint(id), result.Summary, string(keyPointsJSON)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save summary"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"summary":    result.Summary,
			"key_points": result.KeyPoints,
			"cached":     false,
		})
	}
}

// GetArticleSummary 获取文章的 AI 概览
func GetArticleSummary(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		article, err := articleRepo.GetArticleByID(uint(id))
		if err != nil || article.UserID != userID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		var keyPoints []string
		json.Unmarshal([]byte(article.KeyPoints), &keyPoints)

		c.JSON(http.StatusOK, gin.H{
			"summary":    article.Summary,
			"key_points": keyPoints,
		})
	}
}

// Tag handlers
func GetTags(tagRepo *repository.TagRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		tags, err := tagRepo.FindByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tags)
	}
}

func CreateTag(tagRepo *repository.TagRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var req struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if tag already exists
		existing, _ := tagRepo.FindByName(userID, req.Name)
		if existing != nil {
			c.JSON(http.StatusOK, existing)
			return
		}

		tag := &models.Tag{
			Name:   req.Name,
			UserID: userID,
		}

		if err := tagRepo.Create(tag); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, tag)
	}
}

func DeleteTag(tagRepo *repository.TagRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
			return
		}

		tag, err := tagRepo.FindByID(uint(id))
		if err != nil || tag.UserID != userID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
			return
		}

		if err := tagRepo.Delete(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tag deleted"})
	}
}

// Article Tag handlers
func AddArticleTag(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ArticleID uint `json:"article_id" binding:"required"`
			TagID     uint `json:"tag_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := articleRepo.AddTag(req.ArticleID, req.TagID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tag added"})
	}
}

func RemoveArticleTag(articleRepo *repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ArticleID uint `json:"article_id" binding:"required"`
			TagID     uint `json:"tag_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := articleRepo.RemoveTag(req.ArticleID, req.TagID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tag removed"})
	}
}

// Dummy handler for compatibility
func DummyHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}
