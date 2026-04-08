package task

import (
	"ImageMaster/core/types"
)

// TaskUpdater 任务更新器实现
type TaskUpdater struct {
	taskID  string
	manager TaskUpdate
}

type TaskUpdate interface {
	UpdateTask(taskID string, updateFunc func(task *DownloadTask))
	UpdateTaskProgress(taskID string, current, total int)
}

// NewTaskUpdater 创建任务更新器
func NewTaskUpdater(taskID string, manager TaskUpdate) *TaskUpdater {
	return &TaskUpdater{
		taskID:  taskID,
		manager: manager,
	}
}

// UpdateTaskName 更新任务名称
func (tu *TaskUpdater) UpdateTaskName(name string) {
	tu.manager.UpdateTask(tu.taskID, func(task *DownloadTask) {
		task.Name = name
	})
}

// UpdateTaskStatus 更新任务状态
func (tu *TaskUpdater) UpdateTaskStatus(status string, errorMsg string) {
	tu.manager.UpdateTask(tu.taskID, func(task *DownloadTask) {
		task.Status = status
		if errorMsg != "" {
			task.Error = errorMsg
		}
	})
}

// UpdateTaskProgress 更新任务进度
func (tu *TaskUpdater) UpdateTaskProgress(current, total int) {
	tu.manager.UpdateTaskProgress(tu.taskID, current, total)
}

// UpdateTaskProgressWithDetails 更新详细进度信息
func (tu *TaskUpdater) UpdateTaskProgressWithDetails(progress types.ProgressDetails) {
	tu.manager.UpdateTask(tu.taskID, func(task *DownloadTask) {
		task.Progress.Current = progress.Current
		task.Progress.Total = progress.Total
		// 可以在这里扩展DownloadTask结构体来存储更多详细信息
	})
}

// UpdateTaskField 更新任务的特定字段
func (tu *TaskUpdater) UpdateTaskField(field string, value interface{}) {
	tu.manager.UpdateTask(tu.taskID, func(task *DownloadTask) {
		switch field {
		case "name":
			if name, ok := value.(string); ok {
				task.Name = name
			}
		case "savePath":
			if path, ok := value.(string); ok {
				task.SavePath = path
			}
		case "status":
			if status, ok := value.(string); ok {
				task.Status = status
			}
		case "error":
			if errorMsg, ok := value.(string); ok {
				task.Error = errorMsg
			}
		}
	})
}

// UpdateTask 使用函数更新任务
func (tu *TaskUpdater) UpdateTask(updateFunc func(task interface{})) {
	tu.manager.UpdateTask(tu.taskID, func(task *DownloadTask) {
		updateFunc(task)
	})
}
