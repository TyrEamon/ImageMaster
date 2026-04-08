package utils

import "context"

// Semaphore 信号量，用于控制并发数量
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore 创建信号量
func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

// Acquire 获取信号量（阻塞直到获取成功）
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// AcquireWithContext 带上下文的获取，支持取消
func (s *Semaphore) AcquireWithContext(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release 释放信号量
func (s *Semaphore) Release() {
	<-s.ch
}

// Available 获取当前可用数量
func (s *Semaphore) Available() int {
	return cap(s.ch) - len(s.ch)
}

// Capacity 获取总容量
func (s *Semaphore) Capacity() int {
	return cap(s.ch)
}

// Used 获取当前使用数量
func (s *Semaphore) Used() int {
	return len(s.ch)
}

// 默认信号量配置
const (
	DefaultRateLimit = 5 // 默认速率限制并发数
)

// 全局默认信号量实例
var (
	// 速率控制信号量 - 用于替代原有的token_bucket功能（简单并发控制）
	DefaultRateSemaphore = NewSemaphore(DefaultRateLimit)
)
