package parsers

import (
	"fmt"

	"ImageMaster/core/logger"
	"ImageMaster/core/request"
	"ImageMaster/core/types"
)

// SetupRequestClient 设置请求客户端的通用配置
func SetupRequestClient(reqClient *request.Client, downloader types.Downloader) error {
	// 使用下载器的代理配置
	if downloader != nil && downloader.GetProxy() != "" {
		reqClient.SetProxy(downloader.GetProxy())
	}

	return nil
}

// UpdateTaskName 更新任务名称
func UpdateTaskName(downloader types.Downloader, name string) {
	if downloader != nil {
		if taskUpdater := downloader.GetTaskUpdater(); taskUpdater != nil {
			taskUpdater.UpdateTaskName(name)
			logger.Debug("已更新任务名称为: %s", name)
		}
	}
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(downloader types.Downloader, status types.DownloadStatus, message string) {
	if downloader != nil {
		if taskUpdater := downloader.GetTaskUpdater(); taskUpdater != nil {
			taskUpdater.UpdateTaskStatus(string(status), message)
		}
	}
}

// UpdateTaskProgress 更新任务进度
func UpdateTaskProgress(downloader types.Downloader, current, total int) {
	if downloader != nil {
		if taskUpdater := downloader.GetTaskUpdater(); taskUpdater != nil {
			taskUpdater.UpdateTaskProgress(current, total)
		}
	}
}

// BatchDownloadWithProgress 带进度的批量下载
func BatchDownloadWithProgress(downloader types.Downloader, imageURLs, filePaths []string) error {
	totalImages := len(imageURLs)
	logger.Info("已收集 %d 张图片URL，开始下载...", totalImages)

	// 更新任务状态为下载中
	UpdateTaskStatus(downloader, types.StatusDownloading, "")
	UpdateTaskProgress(downloader, 0, totalImages)

	// 批量下载所有图片
	headers := make(map[string]string)
	successImages, err := downloader.BatchDownload(imageURLs, filePaths, headers)
	if err != nil {
		logger.Error("批量下载出错: %v", err)
		return fmt.Errorf("批量下载出错: %w", err)
	}

	logger.Info("下载完成，总共 %d 张图片，成功 %d 张", totalImages, successImages)

	// 更新最终状态
	if successImages == totalImages {
		UpdateTaskStatus(downloader, types.StatusCompleted, "")
	} else {
		failedCount := totalImages - successImages
		UpdateTaskStatus(downloader, types.StatusFailed, fmt.Sprintf("成功 %d 张，失败 %d 张", successImages, failedCount))
	}

	// 如果有图片下载失败，返回错误
	if successImages < totalImages {
		failedCount := totalImages - successImages
		return fmt.Errorf("下载未完全成功，总共 %d 张图片，成功 %d 张，失败 %d 张", totalImages, successImages, failedCount)
	}

	return nil
}

// ValidateDownloader 验证下载器
func ValidateDownloader(downloader types.Downloader, parserName string) error {
	if downloader == nil {
		return fmt.Errorf("未提供下载器")
	}
	logger.Debug("%s解析器使用传入的下载器", parserName)
	return nil
}
