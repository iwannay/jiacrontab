package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var DefaultLogger *log.Logger
var logChan chan *logContent

const (
	levelDebug = "[DEBUG] "
	levelInfo  = "[INFO] "
	levelWarn  = "[WARN] "
	levelError = "[ERROR] "
	levelFatal = "[FATAL] "
)

type logContent struct {
	level   string
	content string
}

func Debug(v ...interface{}) {
	logChan <- &logContent{
		level:   levelDebug,
		content: fmt.Sprintln(v...),
	}
}

func Info(v ...interface{}) {
	logChan <- &logContent{
		level:   levelInfo,
		content: fmt.Sprintln(v...),
	}
}

func Warn(v ...interface{}) {
	logChan <- &logContent{
		level:   levelWarn,
		content: fmt.Sprintln(v...),
	}
}

func Error(v ...interface{}) {
	logChan <- &logContent{
		level:   levelError,
		content: fmt.Sprintln(v...),
	}
}

func Fatal(v ...interface{}) {
	logChan <- &logContent{
		level:   levelFatal,
		content: fmt.Sprintln(v...),
	}
}

func Debugf(format string, v ...interface{}) {
	logChan <- &logContent{
		level:   levelDebug,
		content: fmt.Sprintf(format, v...),
	}
}

func Infof(format string, v ...interface{}) {

	logChan <- &logContent{
		level:   levelInfo,
		content: fmt.Sprintf(format, v...),
	}
}

func Warnf(format string, v ...interface{}) {
	logChan <- &logContent{
		level:   levelWarn,
		content: fmt.Sprintf(format, v...),
	}
}

func Errorf(format string, v ...interface{}) {
	logChan <- &logContent{
		level:   levelError,
		content: fmt.Sprintf(format, v...),
	}
}

func Fatalf(format string, v ...interface{}) {
	logChan <- &logContent{
		level:   levelFatal,
		content: fmt.Sprintf(format, v...),
	}
}

func SetOptput(w io.Writer) {
	DefaultLogger.SetOutput(w)
}

func SetFlags(flag int) {
	DefaultLogger.SetFlags(flag)
}

func init() {
	logChan = make(chan *logContent)
	DefaultLogger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	go func() {
		for {
			select {
			case log := <-logChan:
				DefaultLogger.SetPrefix(log.level)
				switch log.level {
				case levelDebug, levelInfo, levelError:
					DefaultLogger.Output(2, log.content)
				case levelFatal:
					DefaultLogger.Output(2, log.content)
					os.Exit(1)
				}
			default:
			}
		}
	}()
}
