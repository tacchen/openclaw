package schedulers

import (
	"log"
	"time"

	"rss-reader/internal/services"
)

// PushScheduler 推送调度器
type PushScheduler struct {
	pushService  *services.PushService
	dailyTicker  *time.Ticker
	weeklyTicker *time.Ticker
	stopChan     chan struct{}
}

// NewPushScheduler 创建推送调度器
func NewPushScheduler(pushService *services.PushService) *PushScheduler {
	return &PushScheduler{
		pushService: pushService,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动调度器
func (s *PushScheduler) Start() {
	log.Println("Starting push scheduler...")

	// 每分钟检查一次是否到达推送时间
	s.dailyTicker = time.NewTicker(1 * time.Minute)
	go s.dailyScheduler()

	s.weeklyTicker = time.NewTicker(1 * time.Minute)
	go s.weeklyScheduler()

	log.Println("Push scheduler started")
}

// Stop 停止调度器
func (s *PushScheduler) Stop() {
	log.Println("Stopping push scheduler...")

	if s.dailyTicker != nil {
		s.dailyTicker.Stop()
	}
	if s.weeklyTicker != nil {
		s.weeklyTicker.Stop()
	}

	close(s.stopChan)
	log.Println("Push scheduler stopped")
}

// dailyScheduler 每日推送调度器
func (s *PushScheduler) dailyScheduler() {
	for {
		select {
		case <-s.dailyTicker.C:
			now := time.Now()
			// 每分钟检查一次
			if err := s.pushService.ProcessDailyPushes(); err != nil {
				log.Printf("Error processing daily pushes: %v", err)
			} else {
				log.Printf("Daily pushes processed at %s", now.Format("2006-01-02 15:04:05"))
			}
		case <-s.stopChan:
			return
		}
	}
}

// weeklyScheduler 每周推送调度器
func (s *PushScheduler) weeklyScheduler() {
	for {
		select {
		case <-s.weeklyTicker.C:
			now := time.Now()
			// 每分钟检查一次
			if err := s.pushService.ProcessWeeklyPushes(); err != nil {
				log.Printf("Error processing weekly pushes: %v", err)
			} else {
				log.Printf("Weekly pushes processed at %s", now.Format("2006-01-02 15:04:05"))
			}
		case <-s.stopChan:
			return
		}
	}
}
