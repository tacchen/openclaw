package services

import (
	"context"
	"log"
	"net/http"
	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"time"

	"github.com/mmcdole/gofeed"
)

type RSSService struct {
	feedRepo     *repository.FeedRepository
	articleRepo  *repository.ArticleRepository
	PushService  *PushService
}

func NewRSSService(feedRepo *repository.FeedRepository, articleRepo *repository.ArticleRepository, pushService *PushService) *RSSService {
	return &RSSService{
		feedRepo:    feedRepo,
		articleRepo: articleRepo,
		PushService: pushService,
	}
}

func (s *RSSService) FetchAllFeeds() {
	feeds, err := s.feedRepo.FindAll()
	if err != nil {
		log.Printf("Error fetching feeds: %v", err)
		return
	}

	for _, feed := range feeds {
		s.FetchAndSaveArticles(&feed, false)
	}
}

// FetchAndSaveArticles 抓取并保存文章，skipPush 控制是否推送
func (s *RSSService) FetchAndSaveArticles(feed *models.Feed, skipPush bool) error {
	// Create custom HTTP client with User-Agent to avoid being blocked
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	fp := gofeed.NewParser()
	fp.Client = client
	fp.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	parsedFeed, err := fp.ParseURLWithContext(feed.URL, context.Background())
	if err != nil {
		log.Printf("Error parsing feed %s: %v", feed.URL, err)
		return err
	}

	// Update feed title if empty
	if feed.Title == "" && parsedFeed.Title != "" {
		feed.Title = parsedFeed.Title
		if err := s.feedRepo.Update(feed); err != nil {
			log.Printf("Error updating feed title: %v", err)
		}
	}

	// Update last fetch time
	now := time.Now()
	feed.LastFetch = &now
	s.feedRepo.Update(feed)

	// Save articles and collect new ones for immediate push
	var newArticles []models.Article
	oneHourAgo := now.Add(-1 * time.Hour)

	for _, item := range parsedFeed.Items {
		// Skip if already exists
		if s.articleRepo.ExistsByLink(feed.ID, item.Link) {
			continue
		}

		article := &models.Article{
			FeedID:      feed.ID,
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			Content:     item.Content,
			UserID:      feed.UserID,
		}

		if item.PublishedParsed != nil {
			article.PubDate = item.PublishedParsed
		}

		if err := s.articleRepo.Create(article); err != nil {
			log.Printf("Error saving article: %v", err)
			continue
		}

		// 加入新文章列表
		newArticles = append(newArticles, *article)
	}

	// 批量发送即时推送（仅推送最近1小时内的文章，避免历史文章刷屏）
	if len(newArticles) > 0 && !skipPush && s.PushService != nil {
		var recentArticles []models.Article
		for _, article := range newArticles {
			// 只推送最近1小时内的文章
			if article.PubDate != nil && article.PubDate.After(oneHourAgo) {
				recentArticles = append(recentArticles, article)
			}
		}

		// 推送最近文章
		if len(recentArticles) > 0 {
			for i := range recentArticles {
				if err := s.PushService.SendImmediatePush(&recentArticles[i]); err != nil {
					log.Printf("Error sending immediate push: %v", err)
				}
			}
			log.Printf("Pushed %d new articles (recent only) from feed %s", len(recentArticles), feed.Title)
		} else {
			log.Printf("Saved %d articles from feed %s (no recent articles to push)", len(newArticles), feed.Title)
		}
	}

	return nil
}
