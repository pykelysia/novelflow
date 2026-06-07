package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger 日志接口，便于后续替换实现（MQ、多例等）
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
}

// Config 日志配置
type Config struct {
	OutputDir string // 日志输出目录，默认 "logs"
	Level     string // debug | info | warn | error，默认 "info"
	Console   bool   // 是否同时输出到控制台
}

var defaultLogger Logger = &slogLogger{l: slog.Default()}

// Init 初始化全局 logger，应在程序启动时调用一次
func Init(cfg Config) error {
	if cfg.OutputDir == "" {
		cfg.OutputDir = "logs"
	}
	if cfg.Level == "" {
		cfg.Level = "info"
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return err
	}

	level := parseLevel(cfg.Level)
	dw := &dailyWriter{dir: cfg.OutputDir}

	var w io.Writer = dw
	if cfg.Console {
		w = io.MultiWriter(dw, os.Stdout)
	}

	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	defaultLogger = &slogLogger{l: slog.New(handler)}
	return nil
}

// SetLogger 替换全局 logger 实现（用于测试或切换到 MQ 实现）
func SetLogger(l Logger) {
	defaultLogger = l
}

func Debug(msg string, args ...any) { defaultLogger.Debug(msg, args...) }
func Info(msg string, args ...any)  { defaultLogger.Info(msg, args...) }
func Warn(msg string, args ...any)  { defaultLogger.Warn(msg, args...) }
func Error(msg string, args ...any) { defaultLogger.Error(msg, args...) }
func Fatal(msg string, args ...any) { defaultLogger.Fatal(msg, args...) }

// ----- slogLogger 实现 -----

type slogLogger struct {
	l *slog.Logger
}

func (s *slogLogger) Debug(msg string, args ...any) { s.l.Debug(msg, args...) }
func (s *slogLogger) Info(msg string, args ...any)  { s.l.Info(msg, args...) }
func (s *slogLogger) Warn(msg string, args ...any)  { s.l.Warn(msg, args...) }
func (s *slogLogger) Error(msg string, args ...any) { s.l.Error(msg, args...) }
func (s *slogLogger) Fatal(msg string, args ...any) {
	s.l.Error(msg, args...)
	os.Exit(1)
}

// ----- dailyWriter：按日期切换日志文件 -----

type dailyWriter struct {
	mu      sync.Mutex
	dir     string
	current string   // 当前日期 "2006-01-02"
	f       *os.File
}

func (d *dailyWriter) Write(p []byte) (int, error) {
	today := time.Now().Format("2006-01-02")
	d.mu.Lock()
	defer d.mu.Unlock()
	if today != d.current {
		if d.f != nil {
			d.f.Close()
		}
		f, err := os.OpenFile(
			filepath.Join(d.dir, today+".log"),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			return 0, err
		}
		d.f = f
		d.current = today
	}
	return d.f.Write(p)
}

func (d *dailyWriter) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.f != nil {
		return d.f.Close()
	}
	return nil
}

// ----- 工具函数 -----

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
