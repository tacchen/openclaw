package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"rss-reader/internal/config"
	"rss-reader/internal/handlers"
	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"rss-reader/internal/schedulers"
	"rss-reader/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load config
	cfg := config.Load()

	// Log webhook URL (masked)
	if cfg.FeishuWebhookURL != "" {
		log.Printf("Feishu webhook configured: %s", maskWebhookURL(cfg.FeishuWebhookURL))
	}

	// Connect to database
	database, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate existing tables
	if err := database.AutoMigrate(&models.User{}, &models.Feed{}, &models.Article{}, &models.Tag{}, &models.ArticleTag{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Skip PushConfig/PushLog migration temporarily due to GORM issue
	// These tables can be created manually or fixed later
	log.Println("Push config tables migration skipped temporarily")

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	feedRepo := repository.NewFeedRepository(database)
	articleRepo := repository.NewArticleRepository(database)
	tagRepo := repository.NewTagRepository(database)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	
	// Initialize Feishu client (optional)
	var feishuClient *services.FeishuClient
	if cfg.FeishuWebhookURL != "" {
		feishuClient = services.NewFeishuClient(cfg.FeishuWebhookURL)
		log.Printf("Feishu webhook enabled: %s", maskWebhookURL(cfg.FeishuWebhookURL))
	} else {
		log.Println("Feishu webhook disabled (FEISHU_WEBHOOK_URL not set)")
	}
	
	rssService := services.NewRSSService(feedRepo, articleRepo, feishuClient)
	openaiService := services.NewOpenAIService()

	// Initialize Push Service
	var pushService *services.PushService
	if feishuClient != nil {
		pushService = services.NewPushService(database, feishuClient)
		log.Println("Push service initialized")
	}

	if openaiService == nil {
		log.Println("Warning: OPENAI_API_KEY not set, AI summary feature disabled")
	}

	// Setup cron for RSS fetching
	c := cron.New()
	c.AddFunc("@every 30m", func() {
		log.Println("Fetching RSS feeds...")
		rssService.FetchAllFeeds()
	})
	
	// Setup cron for daily summary (每天 9:00 汇总推送）
	if pushService != nil {
		c.AddFunc("0 9 * * *", func() {
			log.Println("Sending daily summary...")
			if err := pushService.SendDailySummary(); err != nil {
				log.Printf("Error sending daily summary: %v", err)
			} else {
				log.Println("Daily summary sent successfully")
			}
		})
	}

	// Initialize and start Push Scheduler
	var pushScheduler *schedulers.PushScheduler
	if pushService != nil {
		pushScheduler = schedulers.NewPushScheduler(pushService)
		pushScheduler.Start()
		defer pushScheduler.Stop()
	}
	
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

			// Push (test endpoint)
			protected.POST("/push/test", handlers.TestPush(pushService))

			// Push Configs
			protected.POST("/push-configs", handlers.CreatePushConfig(pushService))
			protected.GET("/push-configs", handlers.GetPushConfigs(pushService))
			protected.GET("/push-configs/:id", handlers.GetPushConfig(pushService))
			protected.PUT("/push-configs/:id", handlers.UpdatePushConfig(pushService))
			protected.DELETE("/push-configs/:id", handlers.DeletePushConfig(pushService))
			protected.POST("/push-configs/:id/test", handlers.TestPushConfig(pushService))

			// Push Logs and Stats
			protected.GET("/push-logs", handlers.GetPushLogs(pushService))
			protected.GET("/push-configs/:id/stats", handlers.GetPushStats(pushService))
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

// maskWebhookURL 隐藏 webhook URL 的敏感部分
func maskWebhookURL(url string) string {
	if len(url) < 30 {
		return "***" + url[len(url)-10:]
	}
	prefix := strings.LastIndex(url, "/hook/")
	if prefix == -1 {
		return url[:15] + "..." + url[len(url)-10:]
	}
	return url[:prefix+6] + "..." + url[len(url)-10:]
}
