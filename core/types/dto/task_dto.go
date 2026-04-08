package dto

import "time"

// DownloadTaskDTO 面向前端的下载任务数据传输对象
// 与内部 task.DownloadTask 保持必要字段一致，便于前端渲染
// 后续可根据需要裁剪或版本化
type DownloadTaskDTO struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	SavePath     string    `json:"savePath"`
	StartTime    time.Time `json:"startTime"`
	CompleteTime time.Time `json:"completeTime"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Error        string    `json:"error"`
	Name         string    `json:"name"`
	Progress     struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progress"`
}
