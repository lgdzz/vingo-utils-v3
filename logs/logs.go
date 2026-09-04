// *****************************************************************************
// 作者: lgdz
// 创建时间: 2026/9/3
// 描述：
// *****************************************************************************

package logs

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/gookit/rotatefile"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
)

var (
	logger    *slog.Logger
	initOnce  sync.Once
	closeOnce sync.Once

	consoleHandler *dynamicConsoleHandler
)

var (
	buffSize   int    = 10 * 1024         // 10k写一次文件
	maxSize    uint64 = 100 * 1024 * 1024 // 最大日志文件100M
	backupTime uint   = 24 * 15           // 日志保留15天
)

func SetBuffSize(size int) {
	buffSize = size
}

func SetMaxSize(size uint64) {
	maxSize = size
}

func SetBackupTime(day uint) {
	backupTime = day * 24
}

// SetConsoleEnabled 设置是否同时输出控制台
func SetConsoleEnabled(enabled bool) {
	if logger == nil {
		return
	}
	consoleHandler.SetEnabled(enabled)
}

// Init 初始化日志
// filePath 可选，默认：runtime/logs/app.log
func Init(filePath ...string) {
	initOnce.Do(func() {
		logPath := "runtime/logs/app.log"

		if len(filePath) > 0 && filePath[0] != "" {
			logPath = filePath[0]
		}

		if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
			panic(err)
		}

		h := handler.MustRotateFile(
			logPath,
			rotatefile.EveryDay,
			handler.WithBuffSize(buffSize),
			handler.WithMaxSize(maxSize),
			handler.WithBackupTime(backupTime),
		)

		format := slog.NewTextFormatter(
			"[{{datetime}}] [{{channel}}] [{{level}}] [{{caller}}] {{message}} {{data}} {{extra}}\n",
		)
		format.TimeFormat = "2006-01-02 15:04:05"

		h.SetFormatter(format)

		logger = slog.NewWithHandlers(h)

		// 创建动态控制台 Handler
		consoleHandler = newDynamicConsoleHandler()
		// 控制台 Handler 始终注册
		// 是否真正输出由 consoleEnabled 控制
		logger.AddHandler(consoleHandler)

		go waitSignal()
	})
}

type dynamicConsoleHandler struct {
	handler *handler.ConsoleHandler
	enabled atomic.Bool
}

func newDynamicConsoleHandler() *dynamicConsoleHandler {
	h := handler.NewConsoleHandler(slog.AllLevels)
	format := slog.NewTextFormatter("[{{datetime}}] [{{channel}}] [{{level}}] [{{caller}}] {{message}} {{data}} {{extra}}\n")
	format.TimeFormat = "2006-01-02 15:04:05"
	format.EnableColor = true
	h.SetFormatter(format)
	d := &dynamicConsoleHandler{
		handler: h,
	}
	return d
}
func (h *dynamicConsoleHandler) IsHandling(level slog.Level) bool {
	if !h.enabled.Load() {
		return false
	}
	return h.handler.IsHandling(level)
}
func (h *dynamicConsoleHandler) Handle(record *slog.Record) error {
	if !h.enabled.Load() {
		return nil
	}
	return h.handler.Handle(record)
}
func (h *dynamicConsoleHandler) Flush() error {
	return h.handler.Flush()
}
func (h *dynamicConsoleHandler) Close() error {
	return h.handler.Close()
}
func (h *dynamicConsoleHandler) SetEnabled(enabled bool) { h.enabled.Store(enabled) }

func waitSignal() {
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	sig := <-ch

	// 自己的清理逻辑
	Close()

	// 恢复默认处理
	signal.Stop(ch)
	signal.Reset(sig)

	// 重新发送信号
	if s, ok := sig.(syscall.Signal); ok {
		_ = syscall.Kill(syscall.Getpid(), s)
	}
}

// Close 关闭日志
func Close() {
	closeOnce.Do(func() {
		if logger != nil {
			logger.MustClose()
		}
	})
}

func Info(args ...any) {
	if logger == nil {
		return
	}
	logger.Info(args...)
}

func Request(args ...any) {
	if logger == nil {
		return
	}
	logger.Info(args...)
}

func Response(args ...any) {
	if logger == nil {
		return
	}
	logger.Info(args...)
}

func Debug(args ...any) {
	if logger == nil {
		return
	}
	logger.Debug(args...)
}

func Warn(args ...any) {
	if logger == nil {
		return
	}
	logger.Warn(args...)
}

func Error(args ...any) {
	if logger == nil {
		return
	}
	logger.Error(args...)
}

func Fatal(args ...any) {
	if logger == nil {
		return
	}
	logger.Fatal(args...)
}
