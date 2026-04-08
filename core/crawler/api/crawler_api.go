package api

import (
	"context"
	"sync"

	"ImageMaster/core/download"
	"ImageMaster/core/task"
	"ImageMaster/core/types"
	"ImageMaster/core/types/dto"
)

// CrawlerAPI 爬虫API接口
type CrawlerAPI struct {
	taskManager   *task.TaskManager
	configManager types.ConfigProvider
	historyStore  types.HistoryStore // 历史存储
	ctx           context.Context    // Wails上下文
	sync.Mutex
}

// Config API配置
type Config struct {
	TaskManagerConfig task.Config
}

// NewCrawlerAPI 创建爬虫API（构造注入 HistoryStore）
func NewCrawlerAPI(configManager types.ConfigProvider, store types.HistoryStore) *CrawlerAPI {
	// 默认配置
	config := Config{
		TaskManagerConfig: task.Config{
			DownloaderConfig: download.Config{
				RetryCount:  3,
				RetryDelay:  2,
				ShowProcess: true,
			},
		},
	}

	api := &CrawlerAPI{
		taskManager:   task.NewTaskManager(config.TaskManagerConfig, store),
		configManager: configManager,
		historyStore:  store,
	}

	// 设置配置管理器
	api.taskManager.SetConfigManager(configManager)

	return api
}

// SetContext 设置Wails上下文
func (api *CrawlerAPI) SetContext(ctx context.Context) {
	api.ctx = ctx
	// 将上下文传递给任务管理器
	api.taskManager.SetContext(ctx)
}

// StartCrawl 开始爬取网页图片
func (api *CrawlerAPI) StartCrawl(url string) string {
	// 调用任务管理器创建爬取任务
	return api.taskManager.CrawlWebImages(url)
}

// CancelCrawl 取消爬取任务
func (api *CrawlerAPI) CancelCrawl(taskID string) bool {
	return api.taskManager.CancelTask(taskID)
}

// GetAllTasks 获取所有任务
func (api *CrawlerAPI) GetAllTasks() []*task.DownloadTask {
	return api.taskManager.GetAllTasks()
}

// GetActiveTasks 获取活跃任务
func (api *CrawlerAPI) GetActiveTasks() []*task.DownloadTask {
	return api.taskManager.GetActiveTasks()
}

// GetHistoryTasks 获取历史任务（按 TaskManager 的排序）
func (api *CrawlerAPI) GetHistoryTasks() []*dto.DownloadTaskDTO {
	return api.taskManager.GetHistoryTasks()
}

// ClearHistory 清除历史记录
func (api *CrawlerAPI) ClearHistory() {
	if api.historyStore != nil {
		api.historyStore.ClearDownloadHistory()
	}
}

// GetTaskByID 根据ID获取任务
func (api *CrawlerAPI) GetTaskByID(taskID string) *task.DownloadTask {
	return api.taskManager.GetTaskByID(taskID)
}

// GetTaskProgress 获取任务进度
func (api *CrawlerAPI) GetTaskProgress(taskID string) map[string]interface{} {
	task := api.taskManager.GetTaskByID(taskID)
	if task == nil {
		return nil
	}

	return map[string]interface{}{
		"id":      task.ID,
		"status":  task.Status,
		"current": task.Progress.Current,
		"total":   task.Progress.Total,
		"percent": calculatePercent(task.Progress.Current, task.Progress.Total),
	}
}

// 计算百分比
func calculatePercent(current, total int) int {
	if total <= 0 {
		return 0
	}
	return int((float64(current) / float64(total)) * 100)
}
