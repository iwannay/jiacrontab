package logger

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/iwannay/jiaweb/utils/file"
)

type (
	JiaLogger interface {
		SetEnableLog(isLog bool)
		SetEnableConsole(isConsole bool)
		SetLogPath(path string)
		Debug(log string, target string)
		Print(log string, target string)
		Info(log string, target string)
		Warn(log string, target string)
		Error(log string, target string)
	}

	JiaLog struct {
		out           *writer
		enableConsole bool
		enableLog     bool
		logLevel      string
		logPath       string
	}

	logContext struct {
		fileName string
		line     int
		fullPath string
		funcName string
	}
)

const (
	LogLevelDebug = "DEBUG"
	LogLevelPrint = "PRINT"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "Error"
)

var (
	jiaLog JiaLogger

	EnableLog      bool = false
	EnableConsole  bool = true
	DefaultLogPath string
)

func NewJiaLog() *JiaLog {
	l := &JiaLog{}
	l.out = NewWriter(l)
	return l
}

func (l *JiaLog) SetEnableLog(isLog bool) {
	l.enableLog = isLog
}

func (l *JiaLog) SetEnableConsole(isConsole bool) {
	l.enableConsole = isConsole
}

func (l *JiaLog) SetLogPath(path string) {
	l.logPath = path
	if !strings.HasSuffix(l.logPath, "/") {
		l.logPath = l.logPath + "/"
	}
}

func (l *JiaLog) Debug(log, target string) {
	l.out.write(log, target, LogLevelDebug, false)
}
func (l *JiaLog) Print(log, target string) {
	l.out.write(log, target, LogLevelPrint, true)
}

func (l *JiaLog) Info(log, target string) {
	l.out.write(log, target, LogLevelInfo, false)
}

func (l *JiaLog) Warn(log, target string) {
	l.out.write(log, target, LogLevelWarn, false)
}
func (l *JiaLog) Error(log, target string) {
	l.out.write(log, target, LogLevelError, false)
}

func Logger() JiaLogger {
	return jiaLog
}

func setLogger(logger JiaLogger) {
	jiaLog = logger
	jiaLog.SetLogPath(DefaultLogPath)
	jiaLog.SetEnableLog(EnableLog)
	jiaLog.SetEnableConsole(EnableConsole)
}

func SetEnableConsole(isConsole bool) {
	EnableLog = isConsole
	if jiaLog != nil {
		jiaLog.SetEnableConsole(isConsole)
	}
}

func SetEnableLog(isLog bool) {
	EnableLog = isLog
	if jiaLog != nil {
		jiaLog.SetEnableLog(isLog)
	}
}

func SetLogPath(path string) {
	DefaultLogPath = path
	if jiaLog != nil {
		jiaLog.SetLogPath(path)
	}
}

func InitJiaLog() {
	if DefaultLogPath == "" {
		DefaultLogPath = file.GetCurrentDirectory()
	}
	if jiaLog == nil {
		jiaLog = NewJiaLog()
	}

	SetLogPath(DefaultLogPath)
	SetEnableLog(EnableLog)
	SetEnableConsole(EnableConsole)
}

func callerInfo(skip int) (ctx *logContext, err error) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil, errors.New("error  during runtime.Callers")
	}

	funcInfo := runtime.FuncForPC(pc)
	if funcInfo == nil {
		return nil, errors.New("error during runtime.FuncForPC")
	}

	funcName := funcInfo.Name()
	if strings.HasPrefix(funcName, ".") {
		funcName = funcName[strings.Index(funcName, "."):]
	}

	ctx = &logContext{
		funcName: filepath.Base(funcName),
		line:     line,
		fullPath: file,
		fileName: filepath.Base(file),
	}

	return ctx, nil

}
