package main

import (
	"log"
	"net/http"
	"time"

	"rss-reader/internal/config"
	"rss-reader/internal/handlers"
	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"rss-reader/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate - 添加新字段
	db.AutoMigrate(&models.User{}, &models.Feed{}, &models.Article{}, &models.Tag{}, &models.ArticleTag{})

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	feedRepo := repository.NewFeedRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	tagRepo := repository.NewTagRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	rssService := services.NewRSSService(feedRepo, articleRepo)
	openaiService := services.NewOpenAIService()

	if openaiService == nil {
		log.Println("Warning: OPENAI_API_KEY not set, AI summary feature disabled")
	}

	// Setup cron for RSS fetching
	c := cron.New()
	c.AddFunc("@every 30m", func() {
		log.Println("Fetching RSS feeds...")
		rssService.FetchAllFeeds()
	})
	c.Start()

	// Fetch feeds immediately on startup
	go func() {
		time.Sleep(5 * time.Second) // Wait for DB to be ready
		log.Println("Initial RSS feed fetch...")
		rssService.FetchAllFeeds()
	}()

	// Setup Gin router
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve frontend static files
	r.Static("/assets", "./frontend/assets")

	// API routes
	api := r.Group("/api")
	{
		// Auth
		api.POST("/auth/register", handlers.Register(authService))
		api.POST("/auth/login", handlers.Login(authService))

		// Protected routes
		protected := api.Group("/")
		protected.Use(handlers.AuthMiddleware(cfg.JWTSecret))
		{
			// Feeds
			protected.GET("/feeds", handlers.GetFeeds(feedRepo))
			protected.POST("/feeds", handlers.CreateFeed(feedRepo, rssService))
			protected.PUT("/feeds/:id", handlers.UpdateFeed(feedRepo))
			protected.DELETE("/feeds/:id", handlers.DeleteFeed(feedRepo))

			// Articles
			protected.GET("/articles", handlers.GetArticles(articleRepo))
			protected.GET("/articles/search", handlers.SearchArticles(articleRepo))
			protected.GET("/articles/unread-count", handlers.GetUnreadCount(articleRepo))
			protected.PATCH("/articles/:id/read", handlers.MarkArticleRead(articleRepo))
			protected.POST("/articles/mark-all-read", handlers.MarkAllRead(articleRepo))

			// AI Summary
			protected.POST("/articles/:id/summary", handlers.GenerateArticleSummary(articleRepo, openaiService))
			protected.GET("/articles/:id/summary", handlers.GetArticleSummary(articleRepo))

			// Tags
			protected.GET("/tags", handlers.GetTags(tagRepo))
			protected.POST("/tags", handlers.CreateTag(tagRepo))
			protected.DELETE("/tags/:id", handlers.DeleteTag(tagRepo))

			// Article Tags
			protected.POST("/articles/tags", handlers.AddArticleTag(articleRepo))
			protected.DELETE("/articles/tags", handlers.RemoveArticleTag(articleRepo))
		}
	}

	// SPA fallback - 所有非 API 路由都返回 index.html
	r.NoRoute(func(c *gin.Context) {
		// 如果是 API 路径但没匹配到，返回 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		// 其他路径返回 index.html（SPA 路由）
		c.File("./frontend/index.html")
	})

	// Start server
	log.Printf("Server starting on port %s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
