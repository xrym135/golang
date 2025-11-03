package logger

import (
	"log"
	"os"
	"strings"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var (
	currentLevel Level
	lg           = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
)

func InitLogger(level string) {
	switch strings.ToLower(level) {
	case "debug":
		currentLevel = DebugLevel
	case "info":
		currentLevel = InfoLevel
	case "warn":
		currentLevel = WarnLevel
	case "error":
		currentLevel = ErrorLevel
	default:
		currentLevel = InfoLevel
	}
	Info("logger initialized level=%s", level)
}

func Debug(fmtStr string, v ...interface{}) {
	if currentLevel <= DebugLevel {
		lg.Printf("[DEBUG] "+fmtStr, v...)
	}
}
func Info(fmtStr string, v ...interface{}) {
	if currentLevel <= InfoLevel {
		lg.Printf("[INFO] "+fmtStr, v...)
	}
}
func Warn(fmtStr string, v ...interface{}) {
	if currentLevel <= WarnLevel {
		lg.Printf("[WARN] "+fmtStr, v...)
	}
}
func Error(fmtStr string, v ...interface{}) {
	if currentLevel <= ErrorLevel {
		lg.Printf("[ERROR] "+fmtStr, v...)
	}
}
func Fatal(fmtStr string, v ...interface{}) {
	lg.Fatalf("[FATAL] "+fmtStr, v...)
}
