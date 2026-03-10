package utils

import (
	"net/url"
	"strings"
)

// GetFaviconURL 从网站 URL 获取 favicon URL
func GetFaviconURL(siteURL string) string {
	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		return ""
	}
	
	domain := parsedURL.Host
	// 移除端口
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	
	// 使用 Google 的 favicon 服务（稳定可靠）
	return "https://www.google.com/s2/favicons?domain=" + domain + "&sz=32"
}
