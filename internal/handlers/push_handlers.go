package handlers

import (
	"net/http"
	"rss-reader/internal/services"

	"github.com/gin-gonic/gin"
)

// TestPush 测试推送
func TestPush(pushService *services.PushService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if pushService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Push service not initialized"})
			return
		}

		// 手动触发每日汇总推送
		if err := pushService.SendDailySummary(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Test push sent successfully",
		})
	}
}
