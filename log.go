package terma

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Logger provides debug logging to a file.
// Logs are written to terma.log in the current directory.
type Logger struct {
	file    *os.File
	mu      sync.Mutex
	enabled bool
}

var globalLogger *Logger

// InitLogger initializes the global logger.
// Call this at the start of your application to enable logging.
func InitLogger() error {
	f, err := os.OpenFile("terma.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	globalLogger = &Logger{
		file:    f,
		enabled: true,
	}
	Log("Logger initialized")
	return nil
}

// CloseLogger closes the global logger.
func CloseLogger() {
	if globalLogger != nil && globalLogger.file != nil {
		globalLogger.file.Close()
		globalLogger = nil
	}
}

// SetDebugLogging enables or disables debug logging.
// Logging is enabled by default when InitLogger is called.
func SetDebugLogging(enabled bool) {
	if globalLogger != nil {
		globalLogger.mu.Lock()
		globalLogger.enabled = enabled
		globalLogger.mu.Unlock()
	}
}

// Log writes a message to the log file.
func Log(format string, args ...any) {
	if globalLogger == nil || !globalLogger.enabled {
		return
	}
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()

	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(globalLogger.file, "[%s] %s\n", timestamp, msg)
}

// LogWidgetRegistry logs all entries in a widget registry.
func LogWidgetRegistry(registry *WidgetRegistry) {
	if globalLogger == nil || !globalLogger.enabled {
		return
	}
	Log("=== Widget Registry (%d entries) ===", len(registry.entries))
	for i, entry := range registry.entries {
		Log("  [%d] ID=%q Bounds={X:%d Y:%d W:%d H:%d} Type=%T",
			i, entry.ID, entry.Bounds.X, entry.Bounds.Y,
			entry.Bounds.Width, entry.Bounds.Height, entry.Widget)
	}
	Log("=== End Registry ===")
}

