package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 使用 slog 作为底层实现，保留现有全局函数签名
type LogLevel = slog.Level

const (
	DebugLevel LogLevel = slog.LevelDebug
	InfoLevel  LogLevel = slog.LevelInfo
	WarnLevel  LogLevel = slog.LevelWarn
	ErrorLevel LogLevel = slog.LevelError
)

var (
	once     sync.Once
	std      *slog.Logger
	levelVar = new(slog.LevelVar)
	writer   io.Writer
	logFile  string
)

type FileConfig struct {
	Filename    string
	MaxSizeMB   int
	MaxBackups  int
	MaxAgeDays  int
	Compress    bool
	WriteStdout bool
}

func defaultLogDir() string {
	if dir, err := os.UserCacheDir(); err == nil && dir != "" {
		return filepath.Join(dir, "ImageMaster", "logs")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".imagemaster", "logs")
}

// 确保 std 存在（即使未显式 Init）
func ensure() {
	if std != nil {
		return
	}
	once.Do(func() {
		levelVar.Set(slog.LevelInfo)
		h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: levelVar})
		std = slog.New(h).With("app", "ImageMaster")
	})
}

// Init 配置文件输出与滚动
func Init(cfg FileConfig) error {
	// 构造 writer
	if cfg.Filename == "" {
		_ = os.MkdirAll(defaultLogDir(), 0o755)
		cfg.Filename = filepath.Join(defaultLogDir(), "app.log")
	}
	logFile = cfg.Filename

	lj := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    max(1, cfg.MaxSizeMB),
		MaxBackups: max(1, cfg.MaxBackups),
		MaxAge:     max(1, cfg.MaxAgeDays),
		Compress:   cfg.Compress,
	}

	if cfg.WriteStdout {
		writer = io.MultiWriter(os.Stderr, lj)
	} else {
		writer = lj
	}

	levelVar.Set(slog.LevelInfo)
	h := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: levelVar})
	std = slog.New(h).With("app", "ImageMaster")
	return nil
}

func SetLevel(level LogLevel) { ensure(); levelVar.Set(level) }

// 全局便捷方法：保持原签名（格式化字符串）
func Debug(format string, args ...interface{}) { ensure(); std.Debug(fmt.Sprintf(format, args...)) }
func Info(format string, args ...interface{})  { ensure(); std.Info(fmt.Sprintf(format, args...)) }
func Warn(format string, args ...interface{})  { ensure(); std.Warn(fmt.Sprintf(format, args...)) }
func Error(format string, args ...interface{}) { ensure(); std.Error(fmt.Sprintf(format, args...)) }
func Fatal(format string, args ...interface{}) {
	ensure()
	std.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// 结构化扩展
func With(args ...any) *slog.Logger { ensure(); return std.With(args...) }

// Panic 保护
func Recover(label string) {
	if r := recover(); r != nil {
		ensure()
		std.Error("panic recovered", "label", label, "err", r, "stack", string(debug.Stack()))
	}
}

func SafeGo(ctx context.Context, label string, fn func(context.Context)) {
	go func() {
		defer Recover(label)
		fn(ctx)
	}()
}

func LogPath() string { return logFile }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
