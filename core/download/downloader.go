package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ImageMaster/core/request"
	"ImageMaster/core/types"
	"ImageMaster/core/utils"
)

// 下载器并发控制常数
const (
	DefaultDownloadConcurrency = 10 // 默认并发下载数量
)

// Downloader 核心下载器
type Downloader struct {
	reqClient     *request.Client
	retryCount    int
	retryDelay    time.Duration
	showProcess   bool
	configManager types.ConfigProvider
	taskUpdater   types.TaskUpdater // 任务更新器
	semaphore     *utils.Semaphore  // 用于控制并发数量的信号量
	mu            sync.RWMutex
	ctx           context.Context
}

// Config 下载器配置
type Config struct {
	RetryCount  int
	RetryDelay  int // 秒
	ShowProcess bool
}

// NewDownloader 创建新的下载器
func NewDownloader(config Config) *Downloader {
	// 创建用于控制并发数量的信号量
	semaphore := utils.NewSemaphore(DefaultDownloadConcurrency)

	return &Downloader{
		reqClient:   request.NewClient(),
		retryCount:  config.RetryCount,
		retryDelay:  time.Duration(config.RetryDelay) * time.Second,
		showProcess: config.ShowProcess,
		semaphore:   semaphore,
	}
}

// SetConfigManager 设置配置管理器
func (d *Downloader) SetConfigManager(configManager types.ConfigProvider) {
	d.configManager = configManager
	// 将配置管理器传递给请求客户端
	d.reqClient.SetConfigManager(configManager)
}

// SetTaskUpdater 设置任务更新器
func (d *Downloader) SetTaskUpdater(updater types.TaskUpdater) {
	d.taskUpdater = updater
}

// GetTaskUpdater 获取任务更新器
func (d *Downloader) GetTaskUpdater() types.TaskUpdater {
	return d.taskUpdater
}

// GetProxy 获取当前代理设置
func (d *Downloader) GetProxy() string {
	return d.configManager.GetProxy()
}

// GetConfigManager 获取配置管理器
func (d *Downloader) GetConfigManager() interface{} {
	return d.configManager
}

// SetContext 设置上下文并同步到请求客户端
func (d *Downloader) SetContext(ctx context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.ctx = ctx
	if d.reqClient != nil {
		d.reqClient.SetContext(ctx)
	}
}

// DownloadFile 下载文件到指定路径
func (d *Downloader) DownloadFile(url string, filePath string, headers map[string]string) error {
	if d.ctx != nil {
		if err := d.ctx.Err(); err != nil {
			return err
		}
	}
	filePath = utils.NormalizePath(filePath)
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 直接创建最终文件
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	// 执行下载
	success := false
	var lastErr error
	for attempt := 0; attempt <= d.retryCount; attempt++ {
		if d.ctx != nil {
			if err := d.ctx.Err(); err != nil {
				lastErr = err
				break
			}
		}
		if attempt > 0 {
			fmt.Printf("重试下载 %s (第 %d 次)\n", url, attempt)
			time.Sleep(d.retryDelay)
		}

		// 设置请求头
		if headers != nil {
			d.reqClient.SetHeaders(headers)
		}

		// 执行请求
		resp, err := d.reqClient.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("请求失败: %w", err)
			continue
		}

		// 检查状态码
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("状态码错误: %d", resp.StatusCode)
			continue
		}

		// 清空文件内容
		if _, err := out.Seek(0, 0); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("文件定位失败: %w", err)
			continue
		}
		if err := out.Truncate(0); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("清空文件失败: %w", err)
			continue
		}

		// 复制数据
		_, err = io.Copy(out, resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("数据写入失败: %w", err)
			continue
		}

		success = true
		break
	}

	if !success {
		// 下载失败时删除文件
		os.Remove(filePath)
		return fmt.Errorf("下载失败: %w", lastErr)
	}

	return nil
}

// DownloadResult 下载结果
type DownloadResult struct {
	Index   int
	URL     string
	Success bool
	Error   error
}

// BatchDownload 批量下载文件（支持并行下载）
func (d *Downloader) BatchDownload(urls []string, filepaths []string, headers map[string]string) (int, error) {
	total := len(urls)
	if total == 0 {
		return 0, nil
	}

	if len(filepaths) != total {
		return 0, fmt.Errorf("URL和文件路径数量不匹配")
	}

	// 创建结果通道
	resultCh := make(chan DownloadResult, total)
	var wg sync.WaitGroup

	// 启动并行下载任务
	for i, url := range urls {
		wg.Add(1)
		go func(index int, downloadURL, filePath string) {
			defer wg.Done()

			// 获取信号量（支持取消）
			if d.ctx != nil {
				if err := d.semaphore.AcquireWithContext(d.ctx); err != nil {
					resultCh <- DownloadResult{Index: index, URL: downloadURL, Success: false, Error: err}
					return
				}
			} else {
				d.semaphore.Acquire()
			}
			defer d.semaphore.Release() // 完成后释放信号量

			// 取消检查（获得并发名额后再次检查）
			if d.ctx != nil {
				if err := d.ctx.Err(); err != nil {
					resultCh <- DownloadResult{Index: index, URL: downloadURL, Success: false, Error: err}
					return
				}
			}

			// 添加日志验证并发控制
			fmt.Printf("开始下载 [%d/%d]: %s (当前并发: %d/%d)\n",
				index+1, len(urls), downloadURL,
				d.semaphore.Used(), d.semaphore.Capacity())

			// 执行下载
			err := d.DownloadFile(downloadURL, filePath, headers)

			// 发送结果
			resultCh <- DownloadResult{
				Index:   index,
				URL:     downloadURL,
				Success: err == nil,
				Error:   err,
			}
		}(i, url, filepaths[i])
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 收集结果并更新进度
	successCount := 0
	completedCount := 0
	for result := range resultCh {
		completedCount++
		if result.Success {
			successCount++
		} else {
			fmt.Printf("下载失败: %s, 错误: %v\n", result.URL, result.Error)
		}

		// 使用任务更新器更新进度
		if d.taskUpdater != nil {
			d.taskUpdater.UpdateTaskProgress(completedCount, total)
			// 提供更详细的进度信息
			progressDetails := types.ProgressDetails{
				Current:     completedCount,
				Total:       total,
				CurrentItem: fmt.Sprintf("并行下载完成: %s (成功: %d/%d)", result.URL, successCount, completedCount),
				Phase:       "downloading",
				Timestamp:   time.Now(),
			}
			d.taskUpdater.UpdateTaskProgressWithDetails(progressDetails)
		}
	}

	return successCount, nil
}
