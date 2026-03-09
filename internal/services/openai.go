package services

import (
	"bytes"
	"strings"
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

	prompt := fmt.Sprintf(`请分析以下文章并生成摘要和关键要点。

文章标题：%s
文章内容：%s

【重要】你必须严格按照以下JSON格式返回，不要添加任何其他内容、解释或markdown代码块标记：

{"summary":"摘要内容（2-3句话）","key_points":["要点1","要点2","要点3"]}

要求：
1. summary: 一段2-3句话的简洁摘要
2. key_points: 3-5个关键要点的字符串数组
3. 直接返回JSON对象，不要用反引号json包裹
4. 不要添加任何前后缀文字`, req.Title, content)

	// 构建请求
	openaiReq := map[string]interface{}{
		"model": s.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个专业的文章摘要助手。请用简洁的中文生成摘要和要点。你必须严格返回纯JSON格式，不要添加任何markdown标记或额外文字。"},
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
	
	// 尝试去除可能的 markdown 代码块标记
	contentStr = strings.TrimSpace(contentStr)
	
	// 处理 ```json ... ``` 格式
	if strings.HasPrefix(contentStr, "```json") {
		contentStr = strings.TrimPrefix(contentStr, "```json")
	} else if strings.HasPrefix(contentStr, "```") {
		contentStr = strings.TrimPrefix(contentStr, "```")
	}
	if strings.HasSuffix(contentStr, "```") {
		contentStr = strings.TrimSuffix(contentStr, "```")
	}
	contentStr = strings.TrimSpace(contentStr)
	
	// 尝试提取 JSON 对象（处理可能的多余内容）
	if idx := strings.Index(contentStr, "{"); idx > 0 {
		contentStr = contentStr[idx:]
	}
	if idx := strings.LastIndex(contentStr, "}"); idx != -1 && idx < len(contentStr)-1 {
		contentStr = contentStr[:idx+1]
	}
	
	if err := json.Unmarshal([]byte(contentStr), &result); err != nil {
		// 如果解析失败，尝试提取内容
		result.Summary = contentStr
		result.KeyPoints = []string{}
	}

	return &result, nil
}
