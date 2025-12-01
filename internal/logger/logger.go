package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger 结构化日志记录器
type Logger struct {
	mu       sync.Mutex
	level    Level
	format   string // "json" or "text"
	output   io.Writer
	fields   map[string]interface{}
}

// LogEntry 日志条目
type LogEntry struct {
	Time    string                 `json:"time"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

var defaultLogger *Logger

func init() {
	defaultLogger = New("info", "json", os.Stdout)
}

// New 创建新的日志记录器
func New(level, format string, output io.Writer) *Logger {
	l := &Logger{
		format: format,
		output: output,
		fields: make(map[string]interface{}),
	}
	l.SetLevel(level)
	return l
}

// Default 获取默认日志记录器
func Default() *Logger {
	return defaultLogger
}

// SetDefault 设置默认日志记录器
func SetDefault(l *Logger) {
	defaultLogger = l
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level string) {
	switch level {
	case "debug":
		l.level = LevelDebug
	case "info":
		l.level = LevelInfo
	case "warn":
		l.level = LevelWarn
	case "error":
		l.level = LevelError
	default:
		l.level = LevelInfo
	}
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: make(map[string]interface{}),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: make(map[string]interface{}),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	if l.format == "json" {
		entry := LogEntry{
			Time:    time.Now().UTC().Format(time.RFC3339),
			Level:   level.String(),
			Message: formattedMsg,
			Fields:  l.fields,
		}
		data, _ := json.Marshal(entry)
		fmt.Fprintln(l.output, string(data))
	} else {
		fieldsStr := ""
		for k, v := range l.fields {
			fieldsStr += fmt.Sprintf(" %s=%v", k, v)
		}
		fmt.Fprintf(l.output, "%s [%s] %s%s\n",
			time.Now().UTC().Format(time.RFC3339),
			level.String(),
			formattedMsg,
			fieldsStr,
		)
	}
}

// Debug 调试日志
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(LevelDebug, msg, args...)
}

// Info 信息日志
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(LevelInfo, msg, args...)
}

// Warn 警告日志
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(LevelWarn, msg, args...)
}

// Error 错误日志
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(LevelError, msg, args...)
}

// 包级别便捷函数
func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

func WithField(key string, value interface{}) *Logger {
	return defaultLogger.WithField(key, value)
}

func WithFields(fields map[string]interface{}) *Logger {
	return defaultLogger.WithFields(fields)
}
