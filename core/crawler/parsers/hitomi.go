package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"

	"ImageMaster/core/request"
	"ImageMaster/core/types"
)

// HitomiFile 表示 Hitomi 的文件结构
type HitomiFile struct {
	Hash    string `json:"hash"`
	HasWebp int    `json:"haswebp"`
	HasAvif int    `json:"hasavif"`
	Name    string `json:"name"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

// HitomiGalleryInfo 表示 Hitomi 的画廊信息
type HitomiGalleryInfo struct {
	Files []HitomiFile `json:"files"`
	ID    string       `json:"id"`
	Title string       `json:"title"`
}

// HitomiCrawler Hitomi网站爬虫
type HitomiCrawler struct {
	*BaseCrawler
}

// NewHitomiCrawler 创建Hitomi爬虫
func NewHitomiCrawler(reqClient *request.Client) *HitomiCrawler {
	parser := &HitomiParser{}
	crawler := &HitomiCrawler{
		BaseCrawler: NewBaseCrawler(reqClient, parser),
	}
	return crawler
}

// 插件注册
func init() {
	Register(SiteTypeHitomi, func(reqClient *request.Client, cfg types.ConfigProvider) types.ImageCrawler {
		return NewHitomiCrawler(reqClient)
	})
	RegisterHostContains(SiteTypeHitomi, "hitomi.la")
}

// SetDownloader 设置下载器，自动包装为HitomiDownloader
func (c *HitomiCrawler) SetDownloader(dl types.Downloader) {
	// 包装下载器以添加Referer头
	wrappedDownloader := NewHitomiDownloader(dl)
	c.BaseCrawler.SetDownloader(wrappedDownloader)
}

// HitomiParser Hitomi解析器
type HitomiParser struct{}

// GetName 获取解析器名称
func (p *HitomiParser) GetName() string {
	return "Hitomi"
}

// Parse 解析Hitomi页面
func (p *HitomiParser) Parse(reqClient *request.Client, url string) (*ParseResult, error) {
	// 从URL中提取ID
	id, err := p.extractID(url)
	if err != nil {
		return nil, fmt.Errorf("提取ID失败: %w", err)
	}

	// 获取画廊信息
	galleryInfo, err := p.getGalleryInfo(reqClient, id)
	if err != nil {
		return nil, fmt.Errorf("获取画廊信息失败: %w", err)
	}

	// 获取页面标题
	title := galleryInfo.Title

	// 获取随机数gg
	ggScript, err := p.getGGScript(reqClient)
	if err != nil {
		return nil, fmt.Errorf("获取gg脚本失败: %w", err)
	}

	// 生成图片URL列表
	imageURLs, err := p.generateImageURLs(id, galleryInfo, ggScript)
	if err != nil {
		return nil, fmt.Errorf("生成图片URL失败: %w", err)
	}

	// 生成文件路径
	filePaths := make([]string, len(imageURLs))
	for i, file := range galleryInfo.Files {
		ext := "webp"
		if file.HasWebp == 0 {
			// 如果没有webp，使用原始文件扩展名
			parts := strings.Split(file.Name, ".")
			if len(parts) > 1 {
				ext = parts[len(parts)-1]
			} else {
				ext = "jpg"
			}
		}
		filePaths[i] = fmt.Sprintf("%03d.%s", i+1, ext)
	}

	return &ParseResult{
		Name:      title,
		ImageURLs: imageURLs,
		FilePaths: filePaths,
	}, nil
}

// extractID 从URL中提取ID
func (p *HitomiParser) extractID(url string) (string, error) {
	// 匹配模式: -数字.html
	re := regexp.MustCompile(`-(\d+)\.html`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("无法从URL中提取ID: %s", url)
	}
	return matches[1], nil
}

// getGalleryInfo 获取画廊信息
func (p *HitomiParser) getGalleryInfo(reqClient *request.Client, id string) (*HitomiGalleryInfo, error) {
	url := fmt.Sprintf("https://ltn.gold-usergeneratedcontent.net/galleries/%s.js", id)

	resp, err := reqClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyStr := string(body)

	// 提取galleryinfo变量
	re := regexp.MustCompile(`var galleryinfo = (.+);?`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("无法找到galleryinfo变量")
	}

	var galleryInfo HitomiGalleryInfo
	err = json.Unmarshal([]byte(matches[1]), &galleryInfo)
	if err != nil {
		return nil, fmt.Errorf("解析galleryinfo失败: %w", err)
	}

	return &galleryInfo, nil
}

// getGGScript 获取gg脚本
func (p *HitomiParser) getGGScript(reqClient *request.Client) (string, error) {
	url := "https://ltn.gold-usergeneratedcontent.net/gg.js"

	resp, err := reqClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// generateImageURLs 生成图片URL列表
func (p *HitomiParser) generateImageURLs(id string, galleryInfo *HitomiGalleryInfo, ggScript string) ([]string, error) {
	// 创建Otto虚拟机
	vm := otto.New()

	// 执行gg脚本
	_, err := vm.Run(ggScript)
	if err != nil {
		return nil, fmt.Errorf("执行gg脚本失败: %w", err)
	}

	// 注入URL生成函数
	urlGeneratorScript := `
var domain2 = 'gold-usergeneratedcontent.net';

function url_from_url_from_hash(galleryid, image, dir, ext, base) {
	if ('tn' === base) {
		return url_from_url('https://a.' + domain2 + '/' + dir + '/' + real_full_path_from_hash(image.hash) + '.' + ext, base);
	}
	return url_from_url(url_from_hash(galleryid, image, dir, ext), base, dir);
}

function real_full_path_from_hash(hash) {
	return hash.replace(/^.*(..)(.)$/, '$2/$1/' + hash);
}

function url_from_url(url, base, dir) {
	return url.replace(/\/\/..?\.(?:gold-usergeneratedcontent\.net|hitomi\.la)\//, '//' + subdomain_from_url(url, base, dir) + '.' + domain2 + '/');
}

function url_from_hash(galleryid, image, dir, ext) {
	ext = ext || dir || image.name.split('.').pop();
	if (dir === 'webp' || dir === 'avif') {
		dir = '';
	} else {
		dir += '/';
	}

	return 'https://a.' + domain2 + '/' + dir + full_path_from_hash(image.hash) + '.' + ext;
}

function full_path_from_hash(hash) {
	return gg.b + gg.s(hash) + '/' + hash;
}

function subdomain_from_url(url, base, dir) {
	var retval = '';
	if (!base) {
		if (dir === 'webp') {
			retval = 'w';
		} else if (dir === 'avif') {
			retval = 'a';
		}
	}

	var b = 16;

	var r = /\/[0-9a-f]{61}([0-9a-f]{2})([0-9a-f])/;
	var m = r.exec(url);
	if (!m) {
		return retval;
	}

	var g = parseInt(m[2] + m[1], b);
	if (!isNaN(g)) {
		if (base) {
			retval = String.fromCharCode(97 + gg.m(g)) + base;
		} else {
			retval = retval + (1 + gg.m(g));
		}
	}

	return retval;
}
`

	_, err = vm.Run(urlGeneratorScript)
	if err != nil {
		return nil, fmt.Errorf("注入URL生成函数失败: %w", err)
	}

	// 生成图片URL
	var imageURLs []string

	for _, file := range galleryInfo.Files {
		// 将文件信息转换为JavaScript对象
		fileJSON, err := json.Marshal(file)
		if err != nil {
			continue
		}

		script := fmt.Sprintf(`
			var file = %s;
			url_from_url_from_hash('%s', file, 'webp');
		`, string(fileJSON), id)

		result, err := vm.Run(script)
		if err != nil {
			fmt.Printf("生成图片URL失败: %v\n", err)
			continue
		}

		imageURL, err := result.ToString()
		if err != nil {
			continue
		}

		imageURLs = append(imageURLs, imageURL)
	}

	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("没有生成任何图片URL")
	}

	return imageURLs, nil
}

// HitomiDownloader 包装下载器以添加Referer头
type HitomiDownloader struct {
	types.Downloader
}

// NewHitomiDownloader 创建带Referer的下载器
func NewHitomiDownloader(downloader types.Downloader) *HitomiDownloader {
	return &HitomiDownloader{
		Downloader: downloader,
	}
}

// BatchDownload 重写批量下载方法以添加Referer头
func (d *HitomiDownloader) BatchDownload(imageURLs, filePaths []string, headers map[string]string) (int, error) {
	// 添加Referer头
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Referer"] = "https://hitomi.la/"

	// 调用原下载器的BatchDownload方法
	return d.Downloader.BatchDownload(imageURLs, filePaths, headers)
}
