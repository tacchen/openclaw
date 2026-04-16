package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

// InteractiveMessage 卡片消息
type InteractiveMessage struct {
	MsgType string      `json:"msg_type"`
	Card    CardContent `json:"card"`
}

// CardContent 卡片内容
type CardContent struct {
	Config   CardConfig `json:"config"`
	Elements []Element  `json:"elements"`
}

// CardConfig 卡片配置
type CardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode"`
}

// Element 卡片元素
type Element struct {
	Tag    string        `json:"tag"`
	Text   *TextContent  `json:"text,omitempty"`
	Actions []Action     `json:"actions,omitempty"`
}

// TextContent 文本内容
type TextContent struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// Action 操作按钮
type Action struct {
	Tag  string       `json:"tag"`
	Text ButtonText   `json:"text"`
	URL  string       `json:"url"`
	Type string       `json:"type"`
}

// ButtonText 按钮文本
type ButtonText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
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

// SendCardMessage 发送卡片消息
func (c *FeishuClient) SendCardMessage(card *InteractiveMessage) error {
	if c.webhookURL == "" {
		return fmt.Errorf("feishu webhook URL not configured")
	}

	return c.sendMessage(card)
}

// SendArticleCard 发送文章卡片
func (c *FeishuClient) SendArticleCard(title, link, description, feedName string) error {
	if c.webhookURL == "" {
		return fmt.Errorf("feishu webhook URL not configured")
	}

	// 限制描述长度
	maxDescLen := 100
	if len(description) > maxDescLen {
		description = description[:maxDescLen] + "..."
	}

	// 构建卡片内容
	var elements []Element

	// 标题元素
	titleText := fmt.Sprintf("**📰 新文章推送**\n\n")
	titleText += fmt.Sprintf("**标题**：%s\n\n", title)
	if feedName != "" {
		titleText += fmt.Sprintf("**来源**：%s\n\n", feedName)
	}
	if description != "" {
		titleText += fmt.Sprintf("**描述**：%s\n\n", description)
	}

	elements = append(elements, Element{
		Tag: "div",
		Text: &TextContent{
			Tag:     "lark_md",
			Content: titleText,
		},
	})

	// 按钮元素
	elements = append(elements, Element{
		Tag: "action",
		Actions: []Action{
			{
				Tag:  "button",
				Type: "primary",
				Text: ButtonText{
					Tag:     "plain_text",
					Content: "查看详情",
				},
				URL: link,
			},
		},
	})

	card := &InteractiveMessage{
		MsgType: "interactive",
		Card: CardContent{
			Config: CardConfig{
				WideScreenMode: true,
			},
			Elements: elements,
		},
	}

	return c.SendCardMessage(card)
}

// SendArticleMessage 发送文章消息（格式化）- 保留兼容性
func (c *FeishuClient) SendArticleMessage(title, link, description, feedName string) error {
	return c.SendArticleCard(title, link, description, feedName)
}

// SendSummaryCard 发送每日汇总卡片
func (c *FeishuClient) SendSummaryCard(articles []SummaryArticle) error {
	if c.webhookURL == "" {
		return fmt.Errorf("feishu webhook URL not configured")
	}

	// 限制文章数量（飞书卡片有大小限制）
	maxArticles := 10
	if len(articles) > maxArticles {
		articles = articles[:maxArticles]
	}

	// 构建卡片内容
	var elements []Element

	// 标题
	titleText := fmt.Sprintf("**📰 每日文章汇总**\n\n")
	titleText += fmt.Sprintf("共 **%d** 篇文章\n\n", len(articles))

	elements = append(elements, Element{
		Tag: "div",
		Text: &TextContent{
			Tag:     "lark_md",
			Content: titleText,
		},
	})

	// 文章列表
	for i, article := range articles {
		articleText := fmt.Sprintf("**%d. %s**\n", i+1, article.Title)

		if article.FeedName != "" {
			articleText += fmt.Sprintf("来源：%s  ", article.FeedName)
		}

		if article.PubDate != "" {
			articleText += fmt.Sprintf("时间：%s\n", article.PubDate)
		}

		if article.Description != "" {
			maxDescLen := 50
			desc := article.Description
			if len(desc) > maxDescLen {
				desc = desc[:maxDescLen] + "..."
			}
			articleText += fmt.Sprintf("描述：%s\n", desc)
		}

		elements = append(elements, Element{
			Tag: "div",
			Text: &TextContent{
				Tag:     "lark_md",
				Content: articleText,
			},
		})

		// 每篇文章后添加分隔线
		elements = append(elements, Element{
			Tag: "hr",
		})
	}

	card := &InteractiveMessage{
		MsgType: "interactive",
		Card: CardContent{
			Config: CardConfig{
				WideScreenMode: true,
			},
			Elements: elements,
		},
	}

	return c.SendCardMessage(card)
}

// SummaryArticle 汇总文章
type SummaryArticle struct {
	Title       string
	Link        string
	FeedName    string
	PubDate     string
	Description string
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
