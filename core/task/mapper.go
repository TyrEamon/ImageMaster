package task

import "ImageMaster/core/types/dto"

// ToDownloadTaskDTO 将内部 DownloadTask 转换为面向前端的 DTO
func ToDownloadTaskDTO(t *DownloadTask) *dto.DownloadTaskDTO {
	if t == nil {
		return nil
	}
	d := &dto.DownloadTaskDTO{
		ID:           t.ID,
		URL:          t.URL,
		Status:       t.Status,
		SavePath:     t.SavePath,
		StartTime:    t.StartTime,
		CompleteTime: t.CompleteTime,
		UpdatedAt:    t.UpdatedAt,
		Error:        t.Error,
		Name:         t.Name,
	}
	d.Progress.Current = t.Progress.Current
	d.Progress.Total = t.Progress.Total
	return d
}
