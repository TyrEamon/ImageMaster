package crawler

import (
	"context"
	"fmt"
	"net/url"

	"ImageMaster/core/crawler/parsers"
	"ImageMaster/core/logger"
	"ImageMaster/core/request"
	"ImageMaster/core/types"
)

// 站点类型常量已迁移至 parsers 包

// CrawlerFactory 爬虫工厂
type CrawlerFactory struct {
	reqClient     *request.Client
	configManager types.ConfigProvider
	ctx           context.Context
}

// NewCrawlerFactory 创建爬虫工厂
func NewCrawlerFactory() *CrawlerFactory {
	return &CrawlerFactory{
		reqClient: request.NewClient(),
	}
}

// SetConfigManager 设置配置管理器
func (f *CrawlerFactory) SetConfigManager(configManager types.ConfigProvider) {
	f.configManager = configManager

	// 如果配置管理器不为空，设置到请求客户端
	if configManager != nil {
		f.reqClient.SetConfigManager(configManager)

		// 从配置中获取代理设置，并直接应用到请求客户端
		if proxyURL := configManager.GetProxy(); proxyURL != "" {
			logger.Debug("设置代理: %s", proxyURL)
			f.reqClient.SetProxy(proxyURL)
		}
	}
}

// SetContext 设置默认请求上下文
func (f *CrawlerFactory) SetContext(ctx context.Context) {
	f.ctx = ctx
	if f.reqClient != nil {
		f.reqClient.SetContext(ctx)
	}
}

// CreateCrawler 创建特定网站的爬虫
func (f *CrawlerFactory) createCrawler(siteType string) types.ImageCrawler {
	logger.Info("创建爬虫, 类型: %s", siteType)

	// 使用注册表创建，所有爬虫共用同一个配置好的请求客户端
	crawler := parsers.CreateCrawler(siteType, f.reqClient, f.configManager)
	if crawler == nil {
		logger.Warn("创建爬虫失败, 类型: %s", siteType)
	}
	return crawler
}

// DetectSiteType 检测网站类型
func (f *CrawlerFactory) detectSiteType(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return parsers.SiteTypeGeneric
	}

	host := parsedURL.Host

	// 统一由 parsers 层的 host 注册表识别
	return parsers.DetectSiteTypeByHost(host)
}

func (f *CrawlerFactory) Create(rawURL string) (types.ImageCrawler, error) {
	siteType := f.detectSiteType(rawURL)
	crawler := f.createCrawler(siteType)
	if crawler == nil {
		return nil, fmt.Errorf("unsupported site type: %s", siteType)
	}
	return crawler, nil
}
