package history

import (
	"ImageMaster/core/logger"
	"ImageMaster/core/types"
	"ImageMaster/core/types/dto"
)

// 确保 Manager 实现 types.HistoryStore 接口
var _ types.HistoryStore = (*Manager)(nil)

// Manager 历史存储管理器
type Manager struct {
	historyManager *HistoryManager
}

// NewManager 创建历史管理器
func NewManager(appName string) *Manager {
	logger.Debug("Initializing history manager for app: %s", appName)
	return &Manager{historyManager: NewHistoryManager(appName)}
}

// AddDownloadRecord 添加下载记录
func (m *Manager) AddDownloadRecord(task *dto.DownloadTaskDTO) {
	m.historyManager.AddRecord(task)
}

// GetDownloadHistory 获取下载历史（强类型）
func (m *Manager) GetDownloadHistory() []*dto.DownloadTaskDTO {
	return m.historyManager.GetHistory()
}

// ClearDownloadHistory 清除下载历史
func (m *Manager) ClearDownloadHistory() {
	m.historyManager.ClearHistory()
}
