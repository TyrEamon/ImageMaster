package request

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"ImageMaster/core/types"
)

// ProxyManager 代理管理器
type ProxyManager struct {
	configProvider types.ConfigProvider
	currentProxy   string
	mu             sync.RWMutex
}

// NewProxyManager 创建代理管理器
func NewProxyManager(configProvider types.ConfigProvider) *ProxyManager {
	pm := &ProxyManager{
		configProvider: configProvider,
	}

	// 从配置中加载代理设置
	if configProvider != nil {
		pm.currentProxy = configProvider.GetProxy()
	}

	return pm
}

// SetProxy 设置代理
func (pm *ProxyManager) SetProxy(proxyURL string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 验证代理URL格式
	if proxyURL != "" {
		_, err := url.Parse(proxyURL)
		if err != nil {
			return fmt.Errorf("无效的代理URL: %w", err)
		}
	}

	// 更新当前代理
	pm.currentProxy = proxyURL

	// 保存到配置
	if pm.configProvider != nil {
		if configManager, ok := pm.configProvider.(types.ConfigManager); ok {
			configManager.SetProxy(proxyURL)
		}
	}

	fmt.Printf("代理已设置为: %s\n", proxyURL)
	return nil
}

// GetProxy 获取当前代理设置
func (pm *ProxyManager) GetProxy() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.currentProxy
}

// CreateHTTPClient 创建带代理的HTTP客户端
func (pm *ProxyManager) CreateHTTPClient() *http.Client {
	pm.mu.RLock()
	proxyURL := pm.currentProxy
	pm.mu.RUnlock()

	// 如果没有代理，返回默认客户端
	if proxyURL == "" {
		return &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	// 解析代理URL
	proxyURLParsed, err := url.Parse(proxyURL)
	if err != nil {
		fmt.Printf("解析代理URL失败: %v，使用默认客户端\n", err)
		return &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	// 创建带代理的Transport
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURLParsed),
	}

	// 返回带代理的客户端
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// RefreshFromConfig 从配置重新加载代理设置
func (pm *ProxyManager) RefreshFromConfig() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.configProvider != nil {
		pm.currentProxy = pm.configProvider.GetProxy()
		fmt.Printf("从配置重新加载代理: %s\n", pm.currentProxy)
	}
}
