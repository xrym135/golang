package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel 日志级别类型
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

// Logger 日志结构体
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger 创建新的日志实例
func NewLogger(level string) *Logger {
	logLevel := INFO
	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = DEBUG
	case "INFO":
		logLevel = INFO
	case "WARNING":
		logLevel = WARNING
	case "ERROR":
		logLevel = ERROR
	}

	return &Logger{
		level:  logLevel,
		logger: log.New(os.Stdout, "", 0),
	}
}

// Debug 输出调试日志
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.output("DEBUG", format, v...)
	}
}

// Info 输出信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.output("INFO", format, v...)
	}
}

// Warning 输出警告日志
func (l *Logger) Warning(format string, v ...interface{}) {
	if l.level <= WARNING {
		l.output("WARNING", format, v...)
	}
}

// Error 输出错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.output("ERROR", format, v...)
	}
}

// output 输出日志
func (l *Logger) output(level, format string, v ...interface{}) {
	now := time.Now().Format("2006-01-02 15:04:05")

	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	} else {
		// 只显示文件名
		parts := strings.Split(file, "/")
		file = parts[len(parts)-1]
	}

	message := fmt.Sprintf(format, v...)
	logEntry := fmt.Sprintf("[%s] [%s] %s:%d - %s", now, level, file, line, message)

	l.logger.Println(logEntry)
}
