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

const (
	DefaultDownloadConcurrency = 10
)

type Downloader struct {
	reqClient     *request.Client
	retryCount    int
	retryDelay    time.Duration
	showProcess   bool
	configManager types.ConfigProvider
	taskUpdater   types.TaskUpdater
	semaphore     *utils.Semaphore
	mu            sync.RWMutex
	ctx           context.Context
}

type Config struct {
	RetryCount  int
	RetryDelay  int
	ShowProcess bool
}

type DownloadResult struct {
	Index   int
	URL     string
	Success bool
	Error   error
}

type DownloadJob struct {
	URL      string
	FilePath string
	Headers  map[string]string
}

func NewDownloader(config Config) *Downloader {
	semaphore := utils.NewSemaphore(DefaultDownloadConcurrency)

	return &Downloader{
		reqClient:   request.NewClient(),
		retryCount:  config.RetryCount,
		retryDelay:  time.Duration(config.RetryDelay) * time.Second,
		showProcess: config.ShowProcess,
		semaphore:   semaphore,
	}
}

func (d *Downloader) SetConfigManager(configManager types.ConfigProvider) {
	d.configManager = configManager
	d.reqClient.SetConfigManager(configManager)
}

func (d *Downloader) SetTaskUpdater(updater types.TaskUpdater) {
	d.taskUpdater = updater
}

func (d *Downloader) GetTaskUpdater() types.TaskUpdater {
	return d.taskUpdater
}

func (d *Downloader) GetProxy() string {
	if d.configManager == nil {
		return ""
	}
	return d.configManager.GetProxy()
}

func (d *Downloader) GetConfigManager() interface{} {
	return d.configManager
}

func (d *Downloader) SetContext(ctx context.Context) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.ctx = ctx
	if d.reqClient != nil {
		d.reqClient.SetContext(ctx)
	}
}

func (d *Downloader) DownloadFile(url string, filePath string, headers map[string]string) error {
	if d.ctx != nil {
		if err := d.ctx.Err(); err != nil {
			return err
		}
	}

	filePath = utils.NormalizePath(filePath)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

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
			fmt.Printf("retrying download %s (%d)\n", url, attempt)
			time.Sleep(d.retryDelay)
		}

		resp, err := d.reqClient.DoRequest(http.MethodGet, url, nil, headers)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		if _, err := out.Seek(0, 0); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("failed to seek file: %w", err)
			continue
		}
		if err := out.Truncate(0); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("failed to truncate file: %w", err)
			continue
		}

		_, err = io.Copy(out, resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to write data: %w", err)
			continue
		}

		success = true
		break
	}

	if !success {
		_ = os.Remove(filePath)
		return fmt.Errorf("download failed: %w", lastErr)
	}

	return nil
}

func (d *Downloader) BatchDownload(urls []string, filepaths []string, headers map[string]string) (int, error) {
	total := len(urls)
	if total == 0 {
		return 0, nil
	}

	if len(filepaths) != total {
		return 0, fmt.Errorf("url count does not match filepath count")
	}

	jobs := make([]DownloadJob, 0, total)
	for i, downloadURL := range urls {
		jobs = append(jobs, DownloadJob{
			URL:      downloadURL,
			FilePath: filepaths[i],
			Headers:  headers,
		})
	}

	return d.BatchDownloadJobs(jobs)
}

func (d *Downloader) BatchDownloadJobs(jobs []DownloadJob) (int, error) {
	total := len(jobs)
	if total == 0 {
		return 0, nil
	}

	resultCh := make(chan DownloadResult, total)
	var wg sync.WaitGroup

	for i, job := range jobs {
		wg.Add(1)
		go func(index int, currentJob DownloadJob) {
			defer wg.Done()

			if d.ctx != nil {
				if err := d.semaphore.AcquireWithContext(d.ctx); err != nil {
					resultCh <- DownloadResult{Index: index, URL: currentJob.URL, Success: false, Error: err}
					return
				}
			} else {
				d.semaphore.Acquire()
			}
			defer d.semaphore.Release()

			if d.ctx != nil {
				if err := d.ctx.Err(); err != nil {
					resultCh <- DownloadResult{Index: index, URL: currentJob.URL, Success: false, Error: err}
					return
				}
			}

			fmt.Printf(
				"starting download [%d/%d]: %s (concurrency: %d/%d)\n",
				index+1,
				len(jobs),
				currentJob.URL,
				d.semaphore.Used(),
				d.semaphore.Capacity(),
			)

			err := d.DownloadFile(currentJob.URL, currentJob.FilePath, currentJob.Headers)
			resultCh <- DownloadResult{
				Index:   index,
				URL:     currentJob.URL,
				Success: err == nil,
				Error:   err,
			}
		}(i, job)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	successCount := 0
	completedCount := 0
	for result := range resultCh {
		completedCount++
		if result.Success {
			successCount++
		} else {
			fmt.Printf("download failed: %s, error: %v\n", result.URL, result.Error)
		}

		if d.taskUpdater != nil {
			d.taskUpdater.UpdateTaskProgress(completedCount, total)
			progressDetails := types.ProgressDetails{
				Current:     completedCount,
				Total:       total,
				CurrentItem: fmt.Sprintf("completed: %s (%d/%d)", result.URL, successCount, completedCount),
				Phase:       "downloading",
				Timestamp:   time.Now(),
			}
			d.taskUpdater.UpdateTaskProgressWithDetails(progressDetails)
		}
	}

	return successCount, nil
}
