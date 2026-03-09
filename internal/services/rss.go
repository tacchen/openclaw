package services

import (
	"log"
	"rss-reader/internal/models"
	"rss-reader/internal/repository"
	"time"

	"github.com/mmcdole/gofeed"
)

type RSSService struct {
	feedRepo    *repository.FeedRepository
	articleRepo *repository.ArticleRepository
}

func NewRSSService(feedRepo *repository.FeedRepository, articleRepo *repository.ArticleRepository) *RSSService {
	return &RSSService{
		feedRepo:    feedRepo,
		articleRepo: articleRepo,
	}
}

func (s *RSSService) FetchAllFeeds() {
	feeds, err := s.feedRepo.FindAll()
	if err != nil {
		log.Printf("Error fetching feeds: %v", err)
		return
	}

	for _, feed := range feeds {
		s.FetchAndSaveArticles(&feed)
	}
}

func (s *RSSService) FetchAndSaveArticles(feed *models.Feed) error {
	fp := gofeed.NewParser()
	parsedFeed, err := fp.ParseURL(feed.URL)
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

	// Save articles
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
		}
	}

	return nil
}
