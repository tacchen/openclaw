package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type OpenAIService struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

func NewOpenAIService() *OpenAIService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil
	}
	
	// 支持自定义 base URL（兼容其他模型厂商如 DeepSeek、Moonshot 等）
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	
	// 支持自定义模型
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	
	return &OpenAIService{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type SummaryRequest struct {
	Title       string
	Description string
	Content     string
}

type SummaryResponse struct {
	Summary   string   `json:"summary"`
	KeyPoints []string `json:"key_points"`
}

func (s *OpenAIService) GenerateSummary(req SummaryRequest) (*SummaryResponse, error) {
	if s == nil || s.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// 构建提示词
	content := req.Content
	if content == "" {
		content = req.Description
	}

	prompt := fmt.Sprintf(`请分析以下文章并生成：
1. 一段2-3句话的摘要
2. 3-5个关键要点

文章标题：%s
文章内容：%s

请以JSON格式返回，格式如下：
{
  "summary": "摘要内容",
  "key_points": ["要点1", "要点2", "要点3"]
}

只返回JSON，不要有其他内容。`, req.Title, content)

	// 构建请求
	openaiReq := map[string]interface{}{
		"model": s.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个专业的文章摘要助手。请用简洁的中文生成摘要和要点。"},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  1000,
	}

	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, err
	}

	// 发送请求 - 使用自定义 base URL
	url := s.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, err
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	// 解析返回的 JSON
	var result SummaryResponse
	contentStr := openaiResp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(contentStr), &result); err != nil {
		// 如果解析失败，尝试提取内容
		result.Summary = contentStr
		result.KeyPoints = []string{}
	}

	return &result, nil
}
