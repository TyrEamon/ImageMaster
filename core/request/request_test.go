package request

import (
	"io"
	"strings"
	"testing"
)

func TestClientProxy(t *testing.T) {
	client := NewClient()
	proxy := "http://127.0.0.1:7890"
	url := "https://18comic.vip/photo/292986"
	client.SetProxy(proxy)
	resp, err := client.Get(url)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	// 检查是否是 Cloudflare 反爬虫页面
	respLower := strings.ToLower(string(body))
	cloudflareIndicators := []string{
		"just a moment",
	}

	isCloudflare := false
	detectedIndicator := ""
	for _, indicator := range cloudflareIndicators {
		if strings.Contains(respLower, indicator) {
			isCloudflare = true
			detectedIndicator = indicator
			break
		}
	}

	if isCloudflare {
		t.Logf("检测到 Cloudflare 反爬虫页面，特征: %s", detectedIndicator)
		t.Errorf("请求被 Cloudflare 反爬虫系统拦截")
	} else {
		t.Logf("成功绕过反爬虫检测，响应长度: %d", len(body))
	}

	// os.WriteFile("test.html", body, 0644)
}
