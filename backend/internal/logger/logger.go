package logger

import (
	"log"

	"go.uber.org/zap"
)

var (
	// L 是在整个后端使用的全局日志实例
	L *zap.Logger
)

// Init 初始化全局日志。应在启动时调用一次
func Init() {
	var err error
	L, err = zap.NewProduction()
	if err != nil {
		log.Printf("failed to initialize zap logger: %v, falling back to no-op logger", err)
		L = zap.NewNop()
	}
}

// Sync 刷新所有缓冲的日志条目
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}
