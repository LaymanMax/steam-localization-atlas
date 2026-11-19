package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	mu     sync.Mutex
	level  Level
	writer io.Writer
}

var defaultLogger = New(os.Stdout, LevelInfo)

func New(w io.Writer, level Level) *Logger {
	return &Logger{
		writer: w,
		level:  level,
	}
}

func ParseLevel(lvl string) Level {
	switch strings.ToUpper(strings.TrimSpace(lvl)) {
	case "DEBUG":
		return LevelDebug
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func SetDefaultLevel(lvl string) {
	defaultLogger.level = ParseLevel(lvl)
}

func (l *Logger) log(lvl Level, lvlStr string, format string, args ...any) {
	if lvl < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	ts := time.Now().UTC().Format(time.RFC3339)
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.writer, "[%s] [%-5s] %s\n", ts, lvlStr, msg)
}

func (l *Logger) Debug(format string, args ...any) { l.log(LevelDebug, "DEBUG", format, args...) }
func (l *Logger) Info(format string, args ...any)  { l.log(LevelInfo, "INFO", format, args...) }
func (l *Logger) Warn(format string, args ...any)  { l.log(LevelWarn, "WARN", format, args...) }
func (l *Logger) Error(format string, args ...any) { l.log(LevelError, "ERROR", format, args...) }

func Debug(format string, args ...any) { defaultLogger.Debug(format, args...) }
func Info(format string, args ...any)  { defaultLogger.Info(format, args...) }
func Warn(format string, args ...any)  { defaultLogger.Warn(format, args...) }
func Error(format string, args ...any) { defaultLogger.Error(format, args...) }

func Default() *Logger         { return defaultLogger }
func StandardLog() *log.Logger { return log.New(defaultLogger.writer, "", 0) }
