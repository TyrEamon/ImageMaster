package parsers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"ImageMaster/core/request"
	"ImageMaster/core/types"
)

// TelegraphAlbum Telegraph专辑
type TelegraphAlbum struct {
	Name   string
	Images []TelegraphImage
}

// TelegraphImage Telegraph图片
type TelegraphImage struct {
	Name string
	URL  string
}

// TelegraphParser Telegraph解析器实现
type TelegraphParser struct{}

// GetName 获取解析器名称
func (p *TelegraphParser) GetName() string {
	return "Telegraph"
}

// Parse 解析URL获取图片信息
func (p *TelegraphParser) Parse(reqClient *request.Client, url string) (*ParseResult, error) {
	album, err := GetTelegraphAlbum(reqClient, url)
	if err != nil {
		return nil, fmt.Errorf("获取专辑失败: %w", err)
	}

	// 准备批量下载
	var imgURLs []string
	var filePaths []string

	for _, image := range album.Images {
		imgURLs = append(imgURLs, image.URL)
		filePaths = append(filePaths, image.Name)
	}

	return &ParseResult{
		Name:      album.Name,
		ImageURLs: imgURLs,
		FilePaths: filePaths,
	}, nil
}

// GetTelegraphAlbum 获取Telegraph专辑
func GetTelegraphAlbum(reqClient *request.Client, url string) (*TelegraphAlbum, error) {
	resp, err := reqClient.Get(url)
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

	// 获取专辑名称
	albumName := ""
	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		albumName = s.Text()
	})

	if albumName == "" {
		albumName = "Telegraph Album" // 默认名称
	}

	// 获取所有图片
	var images []TelegraphImage
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			url := FormatTelegraphURL(src)

			// 创建图片信息
			image := TelegraphImage{
				Name: fmt.Sprintf("%d.jpg", i),
				URL:  url,
			}

			images = append(images, image)
		}
	})

	return &TelegraphAlbum{
		Name:   albumName,
		Images: images,
	}, nil
}

// FormatTelegraphURL 格式化Telegraph URL
func FormatTelegraphURL(url string) string {
	if strings.HasPrefix(url, "http") {
		return url
	}
	return "https://telegra.ph" + url
}

// TelegraphCrawler Telegraph爬虫
type TelegraphCrawler struct {
	*BaseCrawler
}

// NewTelegraphCrawler 创建新的Telegraph爬虫
func NewTelegraphCrawler(reqClient *request.Client) types.ImageCrawler {
	parser := &TelegraphParser{}
	baseCrawler := NewBaseCrawler(reqClient, parser)
	return &TelegraphCrawler{
		BaseCrawler: baseCrawler,
	}
}

// 插件注册
func init() {
	Register(SiteTypeTelegraph, func(reqClient *request.Client, cfg types.ConfigProvider) types.ImageCrawler {
		return NewTelegraphCrawler(reqClient)
	})
	RegisterHostContains(SiteTypeTelegraph, "telegra.ph", "telegraph.com")
}
