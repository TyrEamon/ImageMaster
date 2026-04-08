package logger

import (
	"os"
	"path/filepath"
	"time"
)

type API struct{}

func NewAPI() *API { return &API{} }

type LogInfo struct {
	Dir         string   `json:"dir"`
	CurrentFile string   `json:"currentFile"`
	SizeBytes   int64    `json:"sizeBytes"`
	Backups     []string `json:"backups"`
	MaxSizeMB   int      `json:"maxSizeMB"`
	MaxBackups  int      `json:"maxBackups"`
	MaxAgeDays  int      `json:"maxAgeDays"`
	Compress    bool     `json:"compress"`
	UpdatedAt   string   `json:"updatedAt"`
}

// GetLogInfo 返回日志目录与当前文件信息
func (a *API) GetLogInfo() (*LogInfo, error) {
	p := LogPath()
	if p == "" {
		// 未初始化时仅返回默认目录
		return &LogInfo{Dir: defaultLogDir()}, nil
	}
	st, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(p)

	pattern := filepath.Base(p) + "*"
	matches, _ := filepath.Glob(filepath.Join(dir, pattern))

	return &LogInfo{
		Dir:         dir,
		CurrentFile: p,
		SizeBytes:   st.Size(),
		Backups:     matches,
		// 这些策略值可按需暴露配置来源，这里给出常用默认
		MaxSizeMB:  50,
		MaxBackups: 5,
		MaxAgeDays: 14,
		Compress:   true,
		UpdatedAt:  st.ModTime().Format(time.RFC3339),
	}, nil
}
