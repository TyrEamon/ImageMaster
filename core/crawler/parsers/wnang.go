package parsers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"

	"ImageMaster/core/request"
	"ImageMaster/core/types"
)

// WnacgAlbum Wnacg专辑
type WnacgAlbum struct {
	Name  string
	Pages []string // 存储所有分页的URL
}

// WnacgParser Wnacg解析器实现
type WnacgParser struct{}

// GetName 获取解析器名称
func (p *WnacgParser) GetName() string {
	return "Wnacg"
}

// Parse 解析URL获取图片信息
func (p *WnacgParser) Parse(reqClient *request.Client, url string) (*ParseResult, error) {
	wnacgAlbum, err := GetWnacgAlbumWithClient(reqClient, url)
	if err != nil {
		return nil, fmt.Errorf("获取专辑失败: %w", err)
	}

	// 批量下载URL和路径
	var imgURLs []string
	var filePaths []string
	var urlMutex sync.Mutex

	// 收集所有漫画页面链接
	var allMangaLinks []string
	for pageIndex, pageURL := range wnacgAlbum.Pages {
		links, err := GetMangaLinksFromPage(reqClient, pageURL)
		if err != nil {
			fmt.Printf("获取分页 %s 的漫画链接失败: %v\n", pageURL, err)
			continue
		}

		// 为每个链接添加页面索引信息
		for linkIndex, link := range links {
			allMangaLinks = append(allMangaLinks, fmt.Sprintf("%d_%d|%s", pageIndex, linkIndex, link))
		}
	}

	totalMangaLinks := len(allMangaLinks)
	fmt.Printf("总共需要处理 %d 个漫画页面\n", totalMangaLinks)

	// 并发处理控制
	var wg sync.WaitGroup

	// 并发处理所有漫画页面链接
	for _, mangaData := range allMangaLinks {
		// 解析页面索引和URL
		parts := strings.SplitN(mangaData, "|", 2)
		if len(parts) != 2 {
			continue
		}
		indexPart := parts[0]
		mangaURL := parts[1]

		wg.Add(1)

		go func(indexPart, mangaURL string) {
			defer wg.Done()

			// 解析漫画页面获取真实图片URL
			imgURL, err := ParseWnacgPageWithClient(reqClient, mangaURL)
			if err != nil {
				fmt.Printf("解析漫画页面失败 %s: %v\n", mangaURL, err)
				return
			}

			// 构建保存文件名
			filename := fmt.Sprintf("%s.jpg", indexPart)

			urlMutex.Lock()
			imgURLs = append(imgURLs, imgURL)
			filePaths = append(filePaths, filename)
			urlMutex.Unlock()

			fmt.Printf("解析完成 %s\n", filename)
		}(indexPart, mangaURL)
	}

	// 等待所有URL收集任务完成
	wg.Wait()

	return &ParseResult{
		Name:      wnacgAlbum.Name,
		ImageURLs: imgURLs,
		FilePaths: filePaths,
	}, nil
}

// GetWnacgAlbumWithClient 获取整个专辑信息，包括所有分页URL
func GetWnacgAlbumWithClient(reqClient *request.Client, url string) (*WnacgAlbum, error) {
	var pageURLs []string
	currentURL := url

	albumName := ""

	// 添加第一页
	pageURLs = append(pageURLs, currentURL)

	// 使用频率限制的请求获取第一页
	resp, err := reqClient.RateLimitedGet(currentURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 读取响应
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	// 获取专辑名称（从页面标题获取）
	albumName = strings.TrimSpace(doc.Find("#bodywrap > h2").Text())
	if albumName == "" {
		albumName = "Unknown Album"
	}

	// 获取所有分页链接
	doc.Find(".paginator a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			// 拼接完整URL
			fullURL := href
			if !strings.HasPrefix(href, "http") {
				fullURL = "https://www.wnacg.com" + href
			}

			// 避免重复添加当前页
			if fullURL != currentURL {
				pageURLs = append(pageURLs, fullURL)
			}
		}
	})

	resp.Body.Close()

	if albumName == "" {
		return nil, fmt.Errorf("无法获取专辑名称")
	}

	// 去重分页URL
	uniqueURLs := make([]string, 0, len(pageURLs))
	seen := make(map[string]bool)
	for _, pageURL := range pageURLs {
		if !seen[pageURL] {
			seen[pageURL] = true
			uniqueURLs = append(uniqueURLs, pageURL)
		}
	}

	return &WnacgAlbum{
		Name:  albumName,
		Pages: uniqueURLs,
	}, nil
}

// GetMangaLinksFromPage 从分页中获取所有漫画页面的链接
func GetMangaLinksFromPage(reqClient *request.Client, pageURL string) ([]string, error) {
	resp, err := reqClient.RateLimitedGet(pageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var links []string
	// 获取 class = cc 的 ul 元素，从 li 的 a 标签中获取每一页漫画的网址
	doc.Find("#bodywrap ul li a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			// 拼接完整URL
			fullURL := href
			if !strings.HasPrefix(href, "http") {
				fullURL = "https://www.wnacg.com" + href
			}
			links = append(links, fullURL)
		}
	})

	return links, nil
}

// ParseWnacgPageWithClient 解析Wnacg漫画页面获取真实图片URL
func ParseWnacgPageWithClient(reqClient *request.Client, link string) (string, error) {
	resp, err := reqClient.RateLimitedGet(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// 获取真实图片URL - 访问每一页漫画的网址，将每一页 id = picarea 的 img 作为图片结果
	imgURL := ""
	doc.Find("#picarea").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			imgURL = src
		}
	})

	if imgURL == "" {
		return "", fmt.Errorf("找不到图片URL")
	}

	// 如果URL是相对路径，转换为绝对路径
	if strings.HasPrefix(imgURL, "//") {
		imgURL = "https:" + imgURL
	} else if strings.HasPrefix(imgURL, "/") {
		imgURL = "https://www.wnacg.com" + imgURL
	}

	return imgURL, nil
}

// WnacgCrawler Wnacg爬虫
type WnacgCrawler struct {
	*BaseCrawler
}

// NewWnacgCrawler 创建新的Wnacg爬虫
func NewWnacgCrawler(reqClient *request.Client) types.ImageCrawler {
	parser := &WnacgParser{}
	baseCrawler := NewBaseCrawler(reqClient, parser)
	return &WnacgCrawler{
		BaseCrawler: baseCrawler,
	}
}

// 插件注册
func init() {
	Register(SiteTypeWnacg, func(reqClient *request.Client, cfg types.ConfigProvider) types.ImageCrawler {
		return NewWnacgCrawler(reqClient)
	})
	RegisterHostContains(SiteTypeWnacg, "wnacg.com")
}
