package task

import (
	"context"
	"sort"
	"sync"
	"time"

	"ImageMaster/core/crawler"
	"ImageMaster/core/download"
	"ImageMaster/core/types"
	"ImageMaster/core/types/dto"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TaskManager 任务管理器
type TaskManager struct {
	tasks         map[string]*DownloadTask        // 所有任务，包括活跃和历史
	activeTasks   map[string]bool                 // 活跃任务集合
	taskCancelMap map[string]chan struct{}        // 任务取消通道
	downloaders   map[string]*download.Downloader // 每个任务对应的下载器实例
	defaultConfig download.Config                 // 默认下载器配置
	mu            sync.RWMutex                    // 并发控制锁
	historyStore  types.HistoryStore              // 历史记录存储
	ctx           context.Context                 // Wails上下文
	configManager types.ConfigProvider            // 配置管理器
}

// Config 任务管理器配置
type Config struct {
	DownloaderConfig download.Config
}

// NewTaskManager 创建任务管理器（构造注入 HistoryStore）
func NewTaskManager(config Config, store types.HistoryStore) *TaskManager {
	return &TaskManager{
		tasks:         make(map[string]*DownloadTask),
		activeTasks:   make(map[string]bool),
		taskCancelMap: make(map[string]chan struct{}),
		downloaders:   make(map[string]*download.Downloader),
		defaultConfig: config.DownloaderConfig,
		historyStore:  store,
	}
}

// SetConfigManager 设置配置管理器
func (tm *TaskManager) SetConfigManager(configManager types.ConfigProvider) {
	tm.configManager = configManager
}

// SetHistoryStore 设置历史记录存储
func (tm *TaskManager) SetHistoryStore(store types.HistoryStore) {
	tm.historyStore = store
}

// SetContext 设置Wails上下文
func (tm *TaskManager) SetContext(ctx context.Context) {
	tm.ctx = ctx
}

// AddTask 添加下载任务并立即开始下载
func (tm *TaskManager) AddTask(url string) *DownloadTask {
	tm.mu.Lock()

	// 创建新任务
	now := time.Now()
	task := &DownloadTask{
		ID:        uuid.New().String(),
		URL:       url,
		Status:    string(types.StatusPending),
		StartTime: now,
		UpdatedAt: now,
	}

	// 初始化进度
	task.Progress.Current = 0
	task.Progress.Total = 0

	// 添加到任务列表
	tm.tasks[task.ID] = task
	tm.activeTasks[task.ID] = true

	// 创建取消通道
	cancelChan := make(chan struct{})
	tm.taskCancelMap[task.ID] = cancelChan

	tm.mu.Unlock()

	// 异步执行下载任务
	go tm.executeTask(task.ID, cancelChan)

	return task
}

// CrawlWebImages 从网页下载图片，返回任务ID
func (tm *TaskManager) CrawlWebImages(url string) string {
	// 添加下载任务
	task := tm.AddTask(url)
	return task.ID
}

// createDownloaderForTask 为任务创建专用的下载器实例
func (tm *TaskManager) createDownloaderForTask(taskID string) *download.Downloader {
	// 创建新的下载器实例
	newDownloader := download.NewDownloader(tm.defaultConfig)

	// 复制配置
	if tm.configManager != nil {
		newDownloader.SetConfigManager(tm.configManager)
	}

	// 创建并设置TaskUpdater
	taskUpdater := NewTaskUpdater(taskID, tm)
	newDownloader.SetTaskUpdater(taskUpdater)

	// 保存到下载器映射
	tm.downloaders[taskID] = newDownloader

	return newDownloader
}

// executeTask 执行下载任务
func (tm *TaskManager) executeTask(taskID string, cancelChan chan struct{}) {
	// 为该任务创建可取消的上下文
	parent := tm.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	// 监听取消通道以触发上下文取消
	go func() {
		<-cancelChan
		cancel()
	}()
	defer func() {
		tm.mu.Lock()
		delete(tm.activeTasks, taskID)
		delete(tm.taskCancelMap, taskID)
		delete(tm.downloaders, taskID)
		tm.mu.Unlock()
	}()

	// 获取任务
	tm.mu.RLock()
	task, exists := tm.tasks[taskID]
	if !exists {
		tm.mu.RUnlock()
		return
	}
	tm.mu.RUnlock()

	// 更新任务状态为下载中
	tm.UpdateTask(taskID, func(task *DownloadTask) {
		task.Status = string(types.StatusParsing)
		task.UpdatedAt = time.Now()
	})

	// 创建下载器
	downloader := tm.createDownloaderForTask(taskID)
	// 传递上下文到下载器
	downloader.SetContext(ctx)

	// 创建爬虫工厂
	crawlerFactory := crawler.NewCrawlerFactory()
	if tm.configManager != nil {
		crawlerFactory.SetConfigManager(tm.configManager)
	}
	// 传递上下文到爬虫工厂
	crawlerFactory.SetContext(ctx)

	// 检测网站类型并创建对应的爬虫
	crawlerInstance, err := crawlerFactory.Create(task.URL)
	if err != nil {
		// 下载失败
		tm.UpdateTask(taskID, func(task *DownloadTask) {
			task.Status = string(types.StatusFailed)
			task.Error = err.Error()
			task.CompleteTime = time.Now()
			task.UpdatedAt = time.Now()
		})
		// 持久化到历史记录
		tm.persistTaskToHistory(taskID)
		return
	}
	crawlerInstance.SetDownloader(downloader)
	// 将上下文传给具体爬虫（若实现）
	if withCtx, ok := crawlerInstance.(interface{ SetContext(context.Context) }); ok {
		withCtx.SetContext(ctx)
	}

	// 设置输出目录
	var outputDir string
	if tm.configManager != nil {
		outputDir = tm.configManager.GetOutputDir()
	} else {
		outputDir = "downloads"
	}

	// 执行爬取
	savePath, err := crawlerInstance.Crawl(task.URL, outputDir)
	if err != nil {
		// 如果是取消，标记为已取消，否则标记失败
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
	} else {
		// 下载成功
		tm.UpdateTask(taskID, func(task *DownloadTask) {
			task.Status = string(types.StatusCompleted)
			task.SavePath = savePath
			task.CompleteTime = time.Now()
			task.UpdatedAt = time.Now()
		})
	}

	// 持久化到历史记录
	tm.persistTaskToHistory(taskID)

	// 发送完成事件到前端
	if tm.ctx != nil {
		runtime.EventsEmit(tm.ctx, "download:completed", map[string]interface{}{
			"taskId": taskID,
			"name":   task.Name,
			"status": task.Status,
		})
	}
}

// persistTaskToHistory 将任务持久化到历史记录
func (tm *TaskManager) persistTaskToHistory(taskID string) {
	tm.mu.RLock()
	task, exists := tm.tasks[taskID]
	tm.mu.RUnlock()
	if !exists || tm.historyStore == nil {
		return
	}
	var dtoTask *dto.DownloadTaskDTO = ToDownloadTaskDTO(task)
	tm.historyStore.AddDownloadRecord(dtoTask)
}

// UpdateTask 更新任务
func (tm *TaskManager) UpdateTask(taskID string, updateFunc func(task *DownloadTask)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		updateFunc(task)
		task.UpdatedAt = time.Now()
	}
}

// UpdateTaskProgress 更新任务进度
func (tm *TaskManager) UpdateTaskProgress(taskID string, current, total int) {
	tm.UpdateTask(taskID, func(task *DownloadTask) {
		task.Progress.Current = current
		task.Progress.Total = total
	})
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if cancelChan, exists := tm.taskCancelMap[taskID]; exists {
		close(cancelChan)
		// 更新任务状态
		if task, exists := tm.tasks[taskID]; exists {
			task.Status = string(types.StatusCancelled)
			task.CompleteTime = time.Now()
			task.UpdatedAt = time.Now()
			// 立即持久化到历史
			if tm.historyStore != nil {
				dtoTask := ToDownloadTaskDTO(task)
				tm.historyStore.AddDownloadRecord(dtoTask)
			}
			// 向前端发事件
			if tm.ctx != nil {
				runtime.EventsEmit(tm.ctx, "download:cancelled", map[string]interface{}{
					"taskId": taskID,
					"name":   task.Name,
					"status": task.Status,
				})
			}
		}
		return true
	}
	return false
}

// GetTaskByID 根据ID获取任务
func (tm *TaskManager) GetTaskByID(taskID string) *DownloadTask {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.tasks[taskID]
}

// GetAllTasks 获取所有任务
func (tm *TaskManager) GetAllTasks() []*DownloadTask {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*DownloadTask, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}

	// 按时间倒序排序
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})

	return tasks
}

// GetActiveTasks 获取活跃任务
func (tm *TaskManager) GetActiveTasks() []*DownloadTask {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*DownloadTask, 0)
	for taskID := range tm.activeTasks {
		if task, exists := tm.tasks[taskID]; exists {
			tasks = append(tasks, task)
		}
	}

	// 按时间倒序排序
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].StartTime.After(tasks[j].StartTime)
	})

	return tasks
}

// GetHistoryTasks 获取历史任务（从磁盘，倒序）
func (tm *TaskManager) GetHistoryTasks() []*dto.DownloadTaskDTO {
	if tm.historyStore == nil {
		return nil
	}
	history := tm.historyStore.GetDownloadHistory()
	// 倒序：优先 completeTime，为零则用 startTime
	sort.Slice(history, func(i, j int) bool {
		ti := history[i].CompleteTime
		if ti.IsZero() {
			ti = history[i].StartTime
		}
		tj := history[j].CompleteTime
		if tj.IsZero() {
			tj = history[j].StartTime
		}
		return ti.After(tj)
	})
	return history
}

// ClearHistory 清除历史记录（磁盘）
func (tm *TaskManager) ClearHistory() {
	if tm.historyStore != nil {
		tm.historyStore.ClearDownloadHistory()
	}
	// 清除非活跃任务
	tm.mu.Lock()
	for id := range tm.tasks {
		if _, active := tm.activeTasks[id]; !active {
			delete(tm.tasks, id)
		}
	}
	tm.mu.Unlock()
}
