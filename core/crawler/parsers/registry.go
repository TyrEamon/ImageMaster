package parsers

import (
	"strings"
	"sync"

	"ImageMaster/core/request"
	"ImageMaster/core/types"
)

// CrawlerConstructor 爬虫构造函数签名
type CrawlerConstructor func(reqClient *request.Client, cfg types.ConfigProvider) types.ImageCrawler

var (
	registryMu      sync.RWMutex
	crawlerRegistry = map[string]CrawlerConstructor{}

	// host 匹配注册表
	hostRegistryMu sync.RWMutex
	hostMatchers   []hostMatcherEntry
)

// 站点类型常量
const (
	SiteTypeEHentai   = "ehentai"
	SiteTypeExHentai  = "exhentai"
	SiteTypeTelegraph = "telegraph"
	SiteTypeWnacg     = "wnacg"
	SiteTypeNhentai   = "nhentai"
	SiteTypeComic18   = "comic18"
	SiteTypeHitomi    = "hitomi"
	SiteTypeGeneric   = "generic"
)

// Register 在注册表中注册站点爬虫构造器
func Register(siteType string, ctor CrawlerConstructor) {
	registryMu.Lock()
	defer registryMu.Unlock()
	crawlerRegistry[siteType] = ctor
}

// CreateCrawler 根据站点类型创建爬虫实例
func CreateCrawler(siteType string, reqClient *request.Client, cfg types.ConfigProvider) types.ImageCrawler {
	registryMu.RLock()
	ctor := crawlerRegistry[siteType]
	registryMu.RUnlock()
	if ctor == nil {
		return nil
	}
	return ctor(reqClient, cfg)
}

// ---- Host 匹配注册与检测 ----

// HostMatcher 用于匹配 host 是否属于某站点
type HostMatcher func(host string) bool

type hostMatcherEntry struct {
	siteType string
	matcher  HostMatcher
}

// RegisterHostMatcher 注册一个自定义 Host 匹配器
func RegisterHostMatcher(siteType string, matcher HostMatcher) {
	hostRegistryMu.Lock()
	defer hostRegistryMu.Unlock()
	hostMatchers = append(hostMatchers, hostMatcherEntry{siteType: siteType, matcher: matcher})
}

// RegisterHostContains 以包含子串的方式注册 Host 规则
func RegisterHostContains(siteType string, substrings ...string) {
	RegisterHostMatcher(siteType, func(host string) bool {
		for _, s := range substrings {
			if s != "" && strings.Contains(host, s) {
				return true
			}
		}
		return false
	})
}

// DetectSiteTypeByHost 根据 host 识别站点类型
func DetectSiteTypeByHost(host string) string {
	hostRegistryMu.RLock()
	defer hostRegistryMu.RUnlock()
	for _, entry := range hostMatchers {
		if entry.matcher != nil && entry.matcher(host) {
			return entry.siteType
		}
	}
	return SiteTypeGeneric
}

// 末尾保留
