package parsers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"

	"ImageMaster/core/logger"
	"ImageMaster/core/request"
	"ImageMaster/core/types"
	"ImageMaster/core/utils"
)

// EHentaiAlbum EH专辑
type EHentaiAlbum struct {
	Name  string
	Pages []string
}

// EHentaiParser EHentai解析器实现
type EHentaiParser struct {
	downloader types.Downloader
	reqClient  *request.Client
	ctx        context.Context

	// 进度跟踪相关属性
	totalImages        int
	completedRealURLs  int
	completedFinalURLs int
	completedLinks     int
	mu                 sync.Mutex
	taskUpdater        types.TaskUpdater
}

// SetContext 注入上下文以支持取消
func (p *EHentaiParser) SetContext(ctx context.Context) {
	p.ctx = ctx
}

// SetDownloader 注入下载器
func (p *EHentaiParser) SetDownloader(dl types.Downloader) {
	p.downloader = dl
	if dl != nil {
		p.taskUpdater = dl.GetTaskUpdater()
	}
}

// GetName 获取解析器名称
func (p *EHentaiParser) GetName() string {
	return "eHentai"
}

// Parse 解析URL获取图片信息
func (p *EHentaiParser) Parse(reqClient *request.Client, url string) (*ParseResult, error) {
	// 解析前取消检查
	if p.ctx != nil {
		if err := p.ctx.Err(); err != nil {
			return nil, err
		}
	}
	// 设置请求客户端和初始化状态
	p.reqClient = reqClient
	if p.downloader != nil {
		p.taskUpdater = p.downloader.GetTaskUpdater()
	}
	if p.taskUpdater != nil {
		p.taskUpdater.UpdateTaskName(p.GetName())
	}

	// 重置进度计数器
	p.completedRealURLs = 0
	p.completedFinalURLs = 0
	p.completedLinks = 0

	// 设置ehentai特殊配置
	err := SetupEHentaiClient(p.reqClient, p.downloader)
	if err != nil {
		return nil, fmt.Errorf("设置EHentai客户端失败: %w", err)
	}

	// 获取专辑信息（耗时操作）
	if p.taskUpdater != nil {
		p.taskUpdater.UpdateTaskName("EHentai - 正在获取专辑信息")
	}
	eHentaiAlbum, err := p.getAlbum(url)
	if err != nil {
		return nil, fmt.Errorf("获取专辑失败: %w", err)
	}

	// 批量下载URL和路径
	var imgURLs []string
	var filePaths []string
	var wg sync.WaitGroup

	// 计算总链接数并设置到结构体属性
	p.totalImages = 0
	for _, page := range eHentaiAlbum.Pages {
		links := ParseLinks(page)
		p.totalImages += len(links)
	}

	// 更新任务名称显示总数
	if p.taskUpdater != nil {
		p.taskUpdater.UpdateTaskName(fmt.Sprintf("EHentai - 正在解析图片链接 (0/%d张)", p.totalImages))
	}

	// 遍历每一页
	for pageIndex, page := range eHentaiAlbum.Pages {
		// 循环取消检查
		if p.ctx != nil {
			if err := p.ctx.Err(); err != nil {
				break
			}
		}
		links := ParseLinks(page)

		// 并发处理每个链接
		for linkIndex, link := range links {
			// 启动前取消检查
			if p.ctx != nil {
				if err := p.ctx.Err(); err != nil {
					break
				}
			}
			wg.Add(1)
			go func(pageIdx, linkIdx int, linkURL string) {
				defer wg.Done()
				// goroutine 内取消检查
				if p.ctx != nil {
					if err := p.ctx.Err(); err != nil {
						return
					}
				}

				// 解析页面获取真实图片URL
				imgURL, err := p.parsePageForImage(linkURL)
				if err != nil {
					logger.Warn("解析页面失败 %s: %v", linkURL, err)
				} else {
					logger.Debug("解析到图片：%s", imgURL)

					// 构建保存文件名
					filename := fmt.Sprintf("%d_%d.jpg", pageIdx, linkIdx)

					// 线程安全地添加到结果中
					p.mu.Lock()
					imgURLs = append(imgURLs, imgURL)
					filePaths = append(filePaths, filename)
					p.mu.Unlock()
				}

				// 更新进度计数器和任务名称
				p.mu.Lock()
				p.completedLinks++
				if p.taskUpdater != nil {
					p.taskUpdater.UpdateTaskName(fmt.Sprintf("EHentai - 解析图片链接进度 (%d/%d张)", p.completedLinks, p.totalImages))
				}
				p.mu.Unlock()
			}(pageIndex, linkIndex, link)
		}
	}

	// 等待所有并发任务完成
	wg.Wait()

	// 解析完成，更新任务名称
	if p.taskUpdater != nil {
		p.taskUpdater.UpdateTaskName("EHentai - 解析完成，准备下载")
	}

	return &ParseResult{
		Name:      eHentaiAlbum.Name,
		ImageURLs: imgURLs,
		FilePaths: filePaths,
	}, nil
}

// getAlbum 获取整个专辑信息
func (p *EHentaiParser) getAlbum(url string) (*EHentaiAlbum, error) {

	// 首先访问第一页获取专辑名称和分页信息
	resp, err := p.reqClient.RateLimitedGet(url)
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
	resp.Body.Close()

	// 获取专辑名称
	albumName := ""
	doc.Find("#gn").Each(func(i int, s *goquery.Selection) {
		albumName = s.Text()
	})

	if albumName == "" {
		return nil, fmt.Errorf("无法获取专辑名称")
	}

	// 获取所有页面URL
	pageURLs := []string{url} // 包含当前页面

	// 获取第一个 .gtb 元素中的所有 td，排除第一个和最后一个
	gtbElement := doc.Find("body > .gtb").First()
	if gtbElement.Length() > 0 {
		tds := gtbElement.Find("td")
		totalTds := tds.Length()

		tds.Each(func(i int, s *goquery.Selection) {
			if i == 0 || i == 1 || i == totalTds-1 {
				return
			}

			// 从td中的a标签获取href
			s.Find("a").Each(func(j int, a *goquery.Selection) {
				if href, exists := a.Attr("href"); exists {
					pageURLs = append(pageURLs, href)
				}
			})
		})
	}

	// 如果只有一页，直接返回当前页面内容
	if len(pageURLs) == 1 {
		html, err := doc.Html()
		if err != nil {
			return nil, err
		}
		return &EHentaiAlbum{
			Name:  albumName,
			Pages: []string{html},
		}, nil
	}

	// 并发访问所有页面
	var pages []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 预分配pages切片
	pages = make([]string, len(pageURLs))

	// 进度计数器
	var completedPages int
	totalPages := len(pageURLs)

	// 更新任务名称显示总页数
	if p.taskUpdater != nil {
		p.taskUpdater.UpdateTaskName(fmt.Sprintf("EHentai - 正在获取专辑页面 (0/%d页)", totalPages))
	}

	for index, pageURL := range pageURLs {
		logger.Debug("访问第%d页, %s", index+1, pageURL)
		// 启动前取消检查
		if p.ctx != nil {
			if err := p.ctx.Err(); err != nil {
				break
			}
		}
		wg.Add(1)
		go func(idx int, pURL string) {
			defer wg.Done()
			// goroutine 内取消检查
			if p.ctx != nil {
				if err := p.ctx.Err(); err != nil {
					return
				}
			}

			pageResp, err := p.reqClient.RateLimitedGet(pURL)
			if err != nil {
				logger.Warn("访问页面失败 %s: %v", pURL, err)
			} else {
				defer pageResp.Body.Close()

				if pageResp.StatusCode != http.StatusOK {
					logger.Warn("页面HTTP状态码错误 %s: %d", pURL, pageResp.StatusCode)
				} else {
					pageDoc, err := goquery.NewDocumentFromReader(pageResp.Body)
					if err != nil {
						logger.Warn("解析页面失败 %s: %v", pURL, err)
					} else {
						html, err := pageDoc.Html()
						if err != nil {
							logger.Warn("获取页面HTML失败 %s: %v", pURL, err)
						} else {
							mu.Lock()
							pages[idx] = html
							mu.Unlock()
						}
					}
				}
			}

			// 更新进度计数器和任务名称
			mu.Lock()
			completedPages++
			if p.taskUpdater != nil {
				p.taskUpdater.UpdateTaskName(fmt.Sprintf("EHentai - 正在获取专辑页面 (%d/%d页)", completedPages, totalPages))
			}
			mu.Unlock()
		}(index, pageURL)
	}

	// 等待所有页面访问完成
	wg.Wait()

	// 过滤掉空的页面
	var validPages []string
	for _, page := range pages {
		if page != "" {
			validPages = append(validPages, page)
		}
	}

	return &EHentaiAlbum{
		Name:  albumName,
		Pages: validPages,
	}, nil
}

// parsePageForImage 解析EH页面获取真实图片URL
func (p *EHentaiParser) parsePageForImage(link string) (string, error) {
	realURL, err := p.getRealURL(link)
	if err != nil {
		return "", fmt.Errorf("获取真实URL失败: %w", err)
	}

	realPage, err := p.parseRealPage(realURL)
	if err != nil {
		return "", fmt.Errorf("解析真实页面失败: %w", err)
	}

	return realPage, nil
}

// getRealURL 获取真实图片URL
func (p *EHentaiParser) getRealURL(link string) (string, error) {
	resp, err := p.reqClient.RateLimitedGet(link)
	logger.Debug("获取真实URL成功...: %s", link)

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

	// 获取img标签的onerror属性
	imgOnError := ""
	doc.Find("#img").Each(func(i int, s *goquery.Selection) {
		if onError, exists := s.Attr("onerror"); exists {
			imgOnError = onError
		}
	})

	if imgOnError == "" {
		return "", fmt.Errorf("找不到图片onerror属性")
	}

	// 提取nl参数
	re := regexp.MustCompile(`nl\('(.+)'\)`)
	matches := re.FindStringSubmatch(imgOnError)
	if len(matches) < 2 {
		return "", fmt.Errorf("无法解析nl参数")
	}

	nl := matches[1]
	realURL := fmt.Sprintf("%s?nl=%s", link, nl)

	// 更新获取真实URL的进度
	p.mu.Lock()
	p.completedRealURLs++
	if p.taskUpdater != nil {
		p.taskUpdater.UpdateTaskName(fmt.Sprintf("EHentai - 已获取%d张图片真实地址 (%d/%d)", p.completedRealURLs, p.completedRealURLs, p.totalImages))
	}
	p.mu.Unlock()

	return realURL, nil
}

// parseRealPage 解析真实页面获取图片URL
func (p *EHentaiParser) parseRealPage(realURL string) (string, error) {
	resp, err := p.reqClient.RateLimitedGet(realURL)
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

	// 获取真实图片URL
	imgURL := ""
	doc.Find("#img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			imgURL = src
		}
	})

	if imgURL == "" {
		return "", fmt.Errorf("找不到图片URL")
	}

	// 解析完成后更新进度
	p.mu.Lock()
	p.completedFinalURLs++
	if p.taskUpdater != nil {
		p.taskUpdater.UpdateTaskName(fmt.Sprintf("EHentai - 已经解析完成%d张图片最终地址 (%d/%d)", p.completedFinalURLs, p.completedFinalURLs, p.totalImages))
	}
	p.mu.Unlock()

	return imgURL, nil
}

// ParseLinks 解析页面中的图片链接
func ParseLinks(body string) []string {
	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil
	}

	var links []string
	doc.Find("#gdt > a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})

	return links
}

// EHentaiCrawler E-Hentai爬虫
type EHentaiCrawler struct {
	*BaseCrawler
}

// NewEHentaiCrawler 创建新的E-Hentai爬虫
func NewEHentaiCrawler(reqClient *request.Client) types.ImageCrawler {
	parser := &EHentaiParser{}
	baseCrawler := NewBaseCrawler(reqClient, parser)
	return &EHentaiCrawler{
		BaseCrawler: baseCrawler,
	}
}

// 插件注册
func init() {
	Register(SiteTypeEHentai, func(reqClient *request.Client, cfg types.ConfigProvider) types.ImageCrawler {
		return NewEHentaiCrawler(reqClient)
	})
	Register(SiteTypeExHentai, func(reqClient *request.Client, cfg types.ConfigProvider) types.ImageCrawler {
		return NewEHentaiCrawler(reqClient)
	})
	// host 规则
	RegisterHostContains(SiteTypeEHentai, "e-hentai.org")
	RegisterHostContains(SiteTypeExHentai, "exhentai.org")
}

// SetupEHentaiClient 设置EHentai特殊的客户端配置
func SetupEHentaiClient(reqClient *request.Client, downloader types.Downloader) error {
	// 先执行通用设置
	if err := SetupRequestClient(reqClient, downloader); err != nil {
		return err
	}

	reqClient.SetSemaphore(utils.NewSemaphore(5))

	// 设置ehentai需要的cookie
	reqClient.AddCookie(&http.Cookie{
		Name:  "nw",
		Value: "1",
	})

	return nil
}
