package history

import (
	"ImageMaster/core/logger"
	"ImageMaster/core/types/dto"
)

// API 历史记录管理API - 对外提供的统一接口（供 Wails 绑定）
type API struct {
	manager *Manager
}

// NewAPI 创建历史 API
func NewAPI(appName string) *API {
	logger.Debug("Creating history API for app: %s", appName)
	return &API{
		manager: NewManager(appName),
	}
}

// GetDownloadHistory 获取下载历史
func (api *API) GetDownloadHistory() []*dto.DownloadTaskDTO {
	return api.manager.GetDownloadHistory()
}

// AddDownloadRecord 添加下载记录
func (api *API) AddDownloadRecord(task *dto.DownloadTaskDTO) {
	api.manager.AddDownloadRecord(task)
}

// ClearDownloadHistory 清除下载历史
func (api *API) ClearDownloadHistory() {
	api.manager.ClearDownloadHistory()
}

// GetStore 获取历史存储（提供给后端注入使用）
func (api *API) GetStore() *Manager {
	return api.manager
}
