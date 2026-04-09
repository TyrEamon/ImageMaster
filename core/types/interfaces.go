package types

import (
	"ImageMaster/core/types/dto"
	"context"
	"time"
)

// TaskUpdater 任务更新器接口
type TaskUpdater interface {
	// UpdateTaskName 更新任务名称
	UpdateTaskName(name string)
	// UpdateTaskStatus 更新任务状态
	UpdateTaskStatus(status string, errorMsg string)
	// UpdateTaskProgress 更新任务进度
	UpdateTaskProgress(current, total int)
	// UpdateTaskProgressWithDetails 更新详细进度信息
	UpdateTaskProgressWithDetails(progress ProgressDetails)
	// UpdateTaskField 更新任务的特定字段
	UpdateTaskField(field string, value interface{})
	// UpdateTask 使用函数更新任务
	UpdateTask(updateFunc func(task interface{}))
}

// ProgressDetails 详细进度信息
type ProgressDetails struct {
	Current     int       `json:"current"`     // 当前进度
	Total       int       `json:"total"`       // 总数
	Speed       string    `json:"speed"`       // 下载速度
	ETA         string    `json:"eta"`         // 预计完成时间
	CurrentItem string    `json:"currentItem"` // 当前处理项目
	Phase       string    `json:"phase"`       // 当前阶段（解析/下载）
	Timestamp   time.Time `json:"timestamp"`   // 时间戳
}

// Downloader 下载器接口
type Downloader interface {
	DownloadFile(url string, filepath string, headers map[string]string) error
	BatchDownload(urls []string, filepaths []string, headers map[string]string) (int, error)
	GetProxy() string
	// GetTaskUpdater 获取任务更新器
	GetTaskUpdater() TaskUpdater
	// SetContext 传入上下文以支持取消
	SetContext(ctx context.Context)
}

// ProgressReporter 进度报告接口
type ProgressReporter interface {
	ReportProgress(current, total int)
}

// ImageCrawler 图片爬虫接口
type ImageCrawler interface {
	Crawl(url string, saveDir string) (string, error)
	CrawlAndSave(url string, savePath string) string
	GetDownloader() Downloader
	SetDownloader(dl Downloader)
}

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	// GetOutputDir 获取输出目录
	GetOutputDir() string

	// GetProxy 获取代理设置
	GetProxy() string
}

// ConfigManager 配置管理接口
type ConfigManager interface {
	GetOutputDir() string
	SetOutputDir() bool
	GetProxy() string
	SetProxy(proxyURL string) bool
	GetBandizipPath() string
	SetBandizipPath(path string) bool
	GetJmCacheDir() string
	SetJmCacheDir(path string) bool
	GetJmCacheRetentionHours() int
	SetJmCacheRetentionHours(hours int) bool
	GetJmCacheSizeLimitMB() int
	SetJmCacheSizeLimitMB(limit int) bool
	GetLibraries() []string
	AddLibrary() bool
	GetActiveLibrary() string
	SetActiveLibrary(library string) bool
}

// HistoryStore 历史记录存储接口（强类型，面向下载任务历史）
// 统一由 history 子系统实现，屏蔽底层持久化细节
// 仅暴露给 task/crawler 等需要读写历史的模块
type HistoryStore interface {
	AddDownloadRecord(task *dto.DownloadTaskDTO)
	GetDownloadHistory() []*dto.DownloadTaskDTO
	ClearDownloadHistory()
}

// DownloadStatus 表示下载任务状态
type DownloadStatus string

const (
	StatusPending     DownloadStatus = "pending"     // 等待下载
	StatusDownloading DownloadStatus = "downloading" // 下载中
	StatusParsing     DownloadStatus = "parsing"     // 解析中
	StatusCompleted   DownloadStatus = "completed"   // 下载完成
	StatusFailed      DownloadStatus = "failed"      // 下载失败
	StatusCancelled   DownloadStatus = "cancelled"   // 已取消
)
