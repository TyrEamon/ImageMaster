package task

import (
	"context"
	"time"

	"ImageMaster/core/jmbridge"
	"ImageMaster/core/logger"
	"ImageMaster/core/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (tm *TaskManager) executeJMTask(ctx context.Context, taskID string, task *DownloadTask, outputDir string) bool {
	if !jmbridge.Supports(task.URL) {
		return false
	}

	if !jmbridge.HelperAvailable() {
		logger.Warn("JM helper not available, fallback to legacy parser for %s", task.URL)
		return false
	}

	proxy := ""
	if tm.configManager != nil {
		proxy = tm.configManager.GetProxy()
	}

	updater := NewTaskUpdater(taskID, tm)
	savePath, err := jmbridge.Download(ctx, updater, task.URL, outputDir, proxy)
	if err != nil {
		tm.UpdateTask(taskID, func(task *DownloadTask) {
			if ctx.Err() == context.Canceled {
				task.Status = string(types.StatusCancelled)
			} else {
				task.Status = string(types.StatusFailed)
				task.Error = err.Error()
			}
			task.CompleteTime = time.Now()
			task.UpdatedAt = time.Now()
		})
		tm.persistTaskToHistory(taskID)
		return true
	}

	tm.UpdateTask(taskID, func(task *DownloadTask) {
		task.Status = string(types.StatusCompleted)
		task.SavePath = savePath
		task.CompleteTime = time.Now()
		task.UpdatedAt = time.Now()
	})
	tm.persistTaskToHistory(taskID)

	if tm.ctx != nil {
		runtime.EventsEmit(tm.ctx, "download:completed", map[string]interface{}{
			"taskId": taskID,
			"name":   task.Name,
			"status": task.Status,
		})
	}

	return true
}
