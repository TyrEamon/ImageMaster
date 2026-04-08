package parsers

import (
	"fmt"
	// "io"
	"net/http"
	// "os"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"ImageMaster/core/request"
	"ImageMaster/core/types"
)

// --------- 废弃，仅作示例 --------------

// Comic18Parser 18comic解析器实现
type Comic18Parser struct{}

// GetName 获取解析器名称
func (p *Comic18Parser) GetName() string {
	return "18Comic"
}

// Parse 解析URL获取图片信息
func (p *Comic18Parser) Parse(reqClient *request.Client, url string) (*ParseResult, error) {
	resp, err := reqClient.Get(url)
	if err != nil {
		fmt.Println("18comic解析器获取URL失败", err)
		return nil, fmt.Errorf("18comic: 网络请求失败: %v", err)
	}
	defer resp.Body.Close()

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, fmt.Errorf("18comic: 读取响应体失败: %v", err)
	// }
	// os.WriteFile("test.html", body, 0644)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("18comic: HTTP状态码错误: %d，请检查URL是否正确或网站是否可访问", resp.StatusCode)
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("18comic解析器解析HTML失败", err)
		return nil, fmt.Errorf("18comic: HTML解析失败: %v，页面可能不是有效的HTML格式", err)
	}

	// 获取专辑名称
	albumName := ""
	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		albumName = s.Text()
	})

	if albumName == "" {
		albumName = "18Comic Album" // 默认名称
	}
	albumName = strings.ReplaceAll(albumName, "/", "_")
	albumName = strings.Trim(albumName, " ")

	// 获取所有图片
	var imgURLs []string
	var filePaths []string
	doc.Find(".scramble-page > img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("data-original"); exists {
			imgURLs = append(imgURLs, src)

			// 从 src 中提取文件扩展名
			ext := path.Ext(src)
			if ext == "" {
				ext = ".webp" // 默认扩展名
			}
			filePaths = append(filePaths, fmt.Sprintf("%d%s", i, ext))
		}
	})

	// 检查是否找到了图片
	if len(imgURLs) == 0 {
		return nil, fmt.Errorf("18comic: 未找到任何图片，可能是：1) URL不正确 2) 页面结构已变化 3) 需要登录才能访问")
	}

	return &ParseResult{
		Name:      albumName,
		ImageURLs: imgURLs,
		FilePaths: filePaths,
	}, nil
}

// Comic18Crawler 18comic爬虫
type Comic18Crawler struct {
	*BaseCrawler
}

// NewComic18Crawler 创建新的18comic爬虫
func NewComic18Crawler(reqClient *request.Client) types.ImageCrawler {
	parser := &Comic18Parser{}
	baseCrawler := NewBaseCrawler(reqClient, parser)
	return &Comic18Crawler{
		BaseCrawler: baseCrawler,
	}
}

// 插件注册
func init() {
	Register(SiteTypeComic18, func(reqClient *request.Client, cfg types.ConfigProvider) types.ImageCrawler {
		return NewComic18Crawler(reqClient)
	})
	RegisterHostContains(SiteTypeComic18, "18comic.vip", "18comic.org")
}
