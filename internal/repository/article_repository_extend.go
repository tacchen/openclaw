package repository

import (
	"time"

	"rss-reader/internal/models"
)

// GetNewArticlesSince 获取指定时间之后的新文章
func (r *ArticleRepository) GetNewArticlesSince(since time.Time) ([]models.Article, error) {
	var articles []models.Article
	err := r.db.Where("created_at > ?", since).
		Where("is_read = ?", false).
		Preload("Feed").
		Order("created_at DESC").
		Find(&articles).Error
	return articles, err
}

// GetNewArticlesSinceByUserID 获取指定用户在指定时间之后的新文章
func (r *ArticleRepository) GetNewArticlesSinceByUserID(userID uint, since time.Time, limit int) ([]models.Article, error) {
	var articles []models.Article
	query := r.db.Where("user_id = ?", userID).
		Where("created_at > ?", since).
		Where("is_read = ?", false).
		Preload("Feed")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&articles).Error
	return articles, err
}

// GetUnreadCountByDate 获取指定日期的未读文章数
func (r *ArticleRepository) GetUnreadCountByDate(date time.Time) (int64, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	var count int64
	err := r.db.Model(&models.Article{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Where("is_read = ?", false).
		Count(&count).Error

	return count, err
}

// GetArticlesByScore 获取所有文章及其重要性分数（用于排序）
// 注意：这个方法需要在应用层计算分数，这里只返回文章
func (r *ArticleRepository) GetArticlesForImportance(since time.Time, limit int) ([]models.Article, error) {
	var articles []models.Article
	query := r.db.Where("created_at > ?", since).
		Preload("Feed")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&articles).Error
	return articles, err
}
