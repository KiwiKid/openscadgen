package pkg

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"log"
	"os"
	"sync"
)

var (
	appLogger *slog.Logger
	logMu     sync.Mutex
)

type stageAttr struct{}

func InitLogger(logFilePath string) error {
	logMu.Lock()
	defer logMu.Unlock()

	if logFilePath == "memory" {
		writer := io.MultiWriter(os.Stdout)
		handler := newColorHandler(writer)
		appLogger = slog.New(handler)
		slog.SetDefault(appLogger)
		logger = log.New(writer, "", log.Ldate|log.Ltime|log.Lshortfile)
		return nil
	}

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	writer := io.MultiWriter(os.Stdout, file)
	handler := newColorHandler(writer)
	appLogger = slog.New(handler)
	slog.SetDefault(appLogger)
	logger = log.New(writer, "", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func Logger() *slog.Logger {
	if appLogger != nil {
		return appLogger
	}
	return slog.Default()
}

func WithStage(stage string) *slog.Logger {
	return Logger().With("stage", stage)
}

func LogDebugf(msg string, args ...any) { Logger().Debug(fmt.Sprintf(msg, args...)) }
func LogInfof(msg string, args ...any)  { Logger().Info(fmt.Sprintf(msg, args...)) }
func LogWarnf(msg string, args ...any)  { Logger().Warn(fmt.Sprintf(msg, args...)) }
func LogErrorf(msg string, args ...any) { Logger().Error(fmt.Sprintf(msg, args...)) }
func LogStagef(stage string, msg string, args ...any) {
	WithStage(stage).Info(fmt.Sprintf(msg, args...))
}

type colorHandler struct {
	mu  sync.Mutex
	out io.Writer
}

func newColorHandler(out io.Writer) slog.Handler { return &colorHandler{out: out} }

func (h *colorHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *colorHandler) WithAttrs(attrs []slog.Attr) slog.Handler      { return h }
func (h *colorHandler) WithGroup(name string) slog.Handler            { return h }

func (h *colorHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	color := "\033[37m"
	switch {
	case r.Level >= slog.LevelError:
		color = "\033[31m"
	case r.Level >= slog.LevelWarn:
		color = "\033[38;5;208m"
	case r.Level >= slog.LevelInfo:
		color = "\033[32m"
	}
	stage := ""
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "stage" {
			stage = a.Value.String()
		}
		return true
	})
	if stage != "" {
		_, _ = fmt.Fprintf(h.out, "%s[%s] [%s] %s\033[0m\n", color, r.Level.String(), stage, r.Message)
		return nil
	}
	_, _ = fmt.Fprintf(h.out, "%s[%s] %s\033[0m\n", color, r.Level.String(), r.Message)
	return nil
}
