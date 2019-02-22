package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var DefaultLogger *log.Logger
var logChan chan *logContent
var logLevel int

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

type logContent struct {
	level   int
	content string
}

func output(log *logContent) {
	content := fmt.Sprintln(logLevels[log.level], log.content)
	switch log.level {
	case LevelDebug, LevelInfo, LevelError:
		if log.level >= logLevel {
			DefaultLogger.Output(3, content)
		}
	case LevelFatal:
		DefaultLogger.Output(3, content)
		os.Exit(1)
	default:
	}
}

func Debug(v ...interface{}) {
	output(&logContent{
		level:   LevelDebug,
		content: fmt.Sprintln(v...),
	})
}

func Info(v ...interface{}) {
	output(&logContent{
		level:   LevelInfo,
		content: fmt.Sprintln(v...),
	})
}

func Warn(v ...interface{}) {
	output(&logContent{
		level:   LevelWarn,
		content: fmt.Sprintln(v...),
	})
}

func Error(v ...interface{}) {
	output(&logContent{
		level:   LevelError,
		content: fmt.Sprintln(v...),
	})
}

func Fatal(v ...interface{}) {
	output(&logContent{
		level:   LevelFatal,
		content: fmt.Sprintln(v...),
	})
}

func Debugf(format string, v ...interface{}) {
	output(&logContent{
		level:   LevelDebug,
		content: fmt.Sprintf(format, v...),
	})
}

func Infof(format string, v ...interface{}) {
	output(&logContent{
		level:   LevelInfo,
		content: fmt.Sprintf(format, v...),
	})
}

func Warnf(format string, v ...interface{}) {
	output(&logContent{
		level:   LevelWarn,
		content: fmt.Sprintf(format, v...),
	})
}

func Errorf(format string, v ...interface{}) {
	output(&logContent{
		level:   LevelError,
		content: fmt.Sprintf(format, v...),
	})
}

func Fatalf(format string, v ...interface{}) {
	output(&logContent{
		level:   LevelFatal,
		content: fmt.Sprintf(format, v...),
	})
}

func SetOptput(w io.Writer) {
	DefaultLogger.SetOutput(w)
}

func SetFlags(flag int) {
	DefaultLogger.SetFlags(flag)
}

func SetLevel(level int) {
	logLevel = level
}

func init() {
	logChan = make(chan *logContent)
	DefaultLogger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}
