package logger

import (
	"log"

	"go.uber.org/zap"
)

var (
	// L is the global logger instance used across the backend.
	L *zap.Logger
)

// Init initializes the global logger. It should be called once at startup.
func Init() {
	var err error
	L, err = zap.NewProduction()
	if err != nil {
		log.Printf("failed to initialize zap logger: %v, falling back to no-op logger", err)
		L = zap.NewNop()
	}
}

// Sync flushes any buffered log entries.
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}
