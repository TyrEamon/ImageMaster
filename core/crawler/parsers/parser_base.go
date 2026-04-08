package parsers

import (
	"context"
	"fmt"
	"path/filepath"

	"ImageMaster/core/logger"
	"ImageMaster/core/request"
	"ImageMaster/core/types"
)

// ParseResult 解析结果
type ParseResult struct {
	Name      string
	ImageURLs []string
	FilePaths []string
}

// Parser 解析器接口
type Parser interface {
	// Parse 解析URL获取图片信息
	Parse(reqClient *request.Client, url string) (*ParseResult, error)
	// GetName 获取解析器名称
	GetName() string
}

// BaseCrawler 基础爬虫结构
type BaseCrawler struct {
	reqClient  *request.Client
	downloader types.Downloader
	parser     Parser
	ctx        context.Context
}

// NewBaseCrawler 创建基础爬虫
func NewBaseCrawler(reqClient *request.Client, parser Parser) *BaseCrawler {
	return &BaseCrawler{
		reqClient: reqClient,
		parser:    parser,
	}
}

// GetDownloader 获取下载器
func (c *BaseCrawler) GetDownloader() types.Downloader {
	return c.downloader
}

// SetDownloader 设置下载器
func (c *BaseCrawler) SetDownloader(dl types.Downloader) {
	c.downloader = dl
	// 如果解析器支持注入下载器，则同步设置
	if da, ok := c.parser.(interface{ SetDownloader(types.Downloader) }); ok {
		da.SetDownloader(dl)
	}
}

// SetContext 设置上下文并传递给请求客户端与下载器
func (c *BaseCrawler) SetContext(ctx context.Context) {
	c.ctx = ctx
	if c.reqClient != nil {
		c.reqClient.SetContext(ctx)
	}
	if c.downloader != nil {
		c.downloader.SetContext(ctx)
	}
	// 解析器若支持上下文，也一并传入
	if withCtx, ok := c.parser.(interface{ SetContext(context.Context) }); ok {
		withCtx.SetContext(ctx)
	}
}

// Crawl 执行爬取
func (c *BaseCrawler) Crawl(url string, savePath string) (string, error) {
	err := c.CrawlWithParser(url, savePath)
	if err != nil {
		return "", err
	}
	return savePath, nil
}

// CrawlAndSave 执行爬取并保存
func (c *BaseCrawler) CrawlAndSave(url string, savePath string) string {
	name := filepath.Base(savePath)
	if name == "" || name == "." {
		name = "download"
	}

	result, err := c.Crawl(url, savePath)
	if err != nil {
		logger.Error("爬取失败: %v", err)
		return ""
	}

	return result
}

// CrawlWithParser 使用解析器执行爬取
func (c *BaseCrawler) CrawlWithParser(url string, savePath string) error {
	logger.Info("下载 %s 内容: %s", c.parser.GetName(), url)

	// 设置请求客户端
	err := SetupRequestClient(c.reqClient, c.downloader)
	if err != nil {
		return fmt.Errorf("设置请求客户端失败: %w", err)
	}

	// 解析前取消检查
	if c.ctx != nil {
		if err := c.ctx.Err(); err != nil {
			return err
		}
	}

	// 解析内容
	result, err := c.parser.Parse(c.reqClient, url)
	if err != nil {
		return fmt.Errorf("解析内容失败: %w", err)
	}

	// 解析后快速取消检查
	if c.ctx != nil {
		if err := c.ctx.Err(); err != nil {
			return err
		}
	}

	// 更新任务名称和状态
	UpdateTaskName(c.downloader, result.Name)
	UpdateTaskStatus(c.downloader, types.StatusParsing, "")

	// 验证下载器
	if c.downloader == nil {
		return fmt.Errorf("未提供下载器")
	}
	logger.Debug("%s解析器使用传入的下载器", c.parser.GetName())

	// 准备下载路径
	contentPath := savePath + "/" + result.Name

	// 准备文件路径
	filePaths := result.FilePaths
	if len(filePaths) == 0 {
		// 如果解析器没有提供文件路径，生成默认路径
		for i := range result.ImageURLs {
			filename := fmt.Sprintf("%03d.jpg", i+1)
			fullPath := fmt.Sprintf("%s/%s", contentPath, filename)
			filePaths = append(filePaths, fullPath)
		}
	} else {
		// 更新文件路径前缀
		for i, path := range filePaths {
			filePaths[i] = fmt.Sprintf("%s/%s", contentPath, filepath.Base(path))
		}
	}

	// 执行批量下载
	return BatchDownloadWithProgress(c.downloader, result.ImageURLs, filePaths)
}
