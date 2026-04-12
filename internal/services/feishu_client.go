package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type FeishuClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewFeishuClient 创建飞书客户端
func NewFeishuClient(webhookURL string) *FeishuClient {
	return &FeishuClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TextMessage 文本消息
type TextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// SendTextMessage 发送文本消息
func (c *FeishuClient) SendTextMessage(text string) error {
	if c.webhookURL == "" {
		return fmt.Errorf("feishu webhook URL not configured")
	}

	msg := TextMessage{
		MsgType: "text",
	}
	msg.Content.Text = text

	return c.sendMessage(msg)
}

// SendArticleMessage 发送文章消息（格式化）
func (c *FeishuClient) SendArticleMessage(title, link, description, feedName string) error {
	if c.webhookURL == "" {
		return fmt.Errorf("feishu webhook URL not configured")
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📰 新文章推送\n\n"))
	builder.WriteString(fmt.Sprintf("标题：%s\n", title))
	builder.WriteString(fmt.Sprintf("来源：%s\n", feedName))

	if description != "" && len(description) > 0 {
		// 限制描述长度
		maxDescLen := 100
		if len(description) > maxDescLen {
			description = description[:maxDescLen] + "..."
		}
		builder.WriteString(fmt.Sprintf("描述：%s\n", description))
	}

	builder.WriteString(fmt.Sprintf("链接：%s", link))

	return c.SendTextMessage(builder.String())
}

// sendMessage 发送消息
func (c *FeishuClient) sendMessage(msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message error: %w", err)
	}

	// 检查消息大小（限制 20KB）
	if len(body) > 20*1024 {
		return fmt.Errorf("message too large: %d bytes (max 20KB)", len(body))
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, _ := json.Marshal(map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":     resp.Status,
	})

	// 记录响应
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu api error: status %d, response: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
