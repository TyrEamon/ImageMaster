package request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http2"

	"ImageMaster/core/types"
	"ImageMaster/core/utils"
)

// Client HTTP客户端封装
type Client struct {
	client         *http.Client
	proxyManager   *ProxyManager
	configManager  types.ConfigProvider
	headers        map[string]string
	cookies        []*http.Cookie
	defaultHeaders map[string]string
	semaphore      *utils.Semaphore
	ctx            context.Context
}

// NewClient 创建新的请求客户端
func NewClient() *Client {
	transport := &uTransport{
		tr1: &http.Transport{},
		tr2: &http2.Transport{},
	}

	return &Client{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		headers: make(map[string]string),
		cookies: make([]*http.Cookie, 0),
		defaultHeaders: map[string]string{
			"accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"accept-language": "en,zh-CN;q=0.9,zh;q=0.8",
			"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
		},
		semaphore: utils.NewSemaphore(10),
	}
}

func (c *Client) SetSemaphore(semaphore *utils.Semaphore) {
	c.semaphore = semaphore
}

// SetConfigManager 设置配置管理器
func (c *Client) SetConfigManager(configManager types.ConfigProvider) {
	c.configManager = configManager
	c.SetProxy(configManager.GetProxy())
}

// SetContext 设置默认请求上下文，用于统一取消/超时控制
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetProxy 设置代理
func (c *Client) SetProxy(proxyURL string) error {
	// 如果没有代理管理器，创建一个
	if c.proxyManager == nil {
		c.proxyManager = NewProxyManager(c.configManager)
	}

	// 设置代理
	err := c.proxyManager.SetProxy(proxyURL)
	if err != nil {
		return err
	}

	// 直接设置uTransport的代理，避免替换整个Transport
	if uTrans, ok := c.client.Transport.(*uTransport); ok {
		if proxyURL != "" {
			proxyURLParsed, parseErr := url.Parse(proxyURL)
			if parseErr != nil {
				return fmt.Errorf("解析代理URL失败: %w", parseErr)
			}
			uTrans.tr1.Proxy = http.ProxyURL(proxyURLParsed)
		} else {
			uTrans.tr1.Proxy = nil
		}
	}

	return nil
}

// GetProxy 获取当前代理设置
func (c *Client) GetProxy() string {
	if c.proxyManager == nil {
		return ""
	}
	return c.proxyManager.GetProxy()
}

// SetHeader 设置请求头
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// SetHeaders 批量设置请求头
func (c *Client) SetHeaders(headers map[string]string) {
	for key, value := range headers {
		c.headers[key] = value
	}
}

// AddCookie 添加Cookie
func (c *Client) AddCookie(cookie *http.Cookie) {
	c.cookies = append(c.cookies, cookie)
}

// ClearCookies 清除所有Cookie
func (c *Client) ClearCookies() {
	c.cookies = make([]*http.Cookie, 0)
}

// Get 发送GET请求
func (c *Client) Get(url string) (*http.Response, error) {
	return c.DoRequest("GET", url, nil, nil)
}

// GetWithContext 发送带上下文的GET请求
func (c *Client) GetWithContext(ctx context.Context, url string) (*http.Response, error) {
	return c.DoRequestWithContext(ctx, "GET", url, nil, nil)
}

// Post 发送POST请求
func (c *Client) Post(url string, body io.Reader, contentType string) (*http.Response, error) {
	headers := map[string]string{
		"Content-Type": contentType,
	}
	return c.DoRequest("POST", url, body, headers)
}

// Head 发送HEAD请求，测试访问性
func (c *Client) Head(url string) (*http.Response, error) {
	return c.DoRequest("HEAD", url, nil, nil)
}

// PostWithContext 发送带上下文的POST请求
func (c *Client) PostWithContext(ctx context.Context, url string, body io.Reader, contentType string) (*http.Response, error) {
	headers := map[string]string{
		"Content-Type": contentType,
	}
	return c.DoRequestWithContext(ctx, "POST", url, body, headers)
}

// DoRequest 执行HTTP请求
func (c *Client) DoRequest(method, url string, body io.Reader, extraHeaders map[string]string) (*http.Response, error) {
	// 尝试从配置中应用代理（如果尚未设置代理且配置管理器存在）
	if c.proxyManager == nil && c.configManager != nil {
		c.proxyManager = NewProxyManager(c.configManager)
		// c.proxyManager.ApplyToClient(c.client)
	}

	// 创建请求
	var req *http.Request
	var err error
	if c.ctx != nil {
		req, err = http.NewRequestWithContext(c.ctx, method, url, body)
	} else {
		req, err = http.NewRequest(method, url, body)
	}
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认头部
	for key, value := range c.defaultHeaders {
		req.Header.Set(key, value)
	}

	// 应用客户端的通用头部
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 应用额外的请求头
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	// 应用Cookie
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	// 执行请求
	return c.client.Do(req)
}

// DoRequestWithContext 执行带上下文的HTTP请求
func (c *Client) DoRequestWithContext(ctx context.Context, method, url string, body io.Reader, extraHeaders map[string]string) (*http.Response, error) {
	// 尝试从配置中应用代理（如果尚未设置代理且配置管理器存在）
	if c.proxyManager == nil && c.configManager != nil {
		c.proxyManager = NewProxyManager(c.configManager)
		// c.proxyManager.ApplyToClient(c.client)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认头部
	for key, value := range c.defaultHeaders {
		req.Header.Set(key, value)
	}

	// 应用客户端的通用头部
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 应用额外的请求头
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	// 应用Cookie
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	// 执行请求
	return c.client.Do(req)
}

// GetHTTPClient 获取底层HTTP客户端
func (c *Client) GetHTTPClient() *http.Client {
	return c.client
}

// RateLimitedGet 使用默认速率信号量的GET请求（替代原token_bucket功能）
func (c *Client) RateLimitedGet(url string) (*http.Response, error) {
	if c.ctx != nil {
		if err := c.semaphore.AcquireWithContext(c.ctx); err != nil {
			return nil, err
		}
	} else {
		c.semaphore.Acquire()
	}
	defer c.semaphore.Release()
	return c.Get(url)
}
