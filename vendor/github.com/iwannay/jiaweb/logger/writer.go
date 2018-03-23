package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/iwannay/jiaweb/utils/file"
)

type (
	writer struct {
		writeChan chan chanLog
		jiaLog    *JiaLog
	}

	chanLog struct {
		Content   string
		LogTarget string
		LogLevel  string
		isRaw     bool
		logCtx    *logContext
	}
)

const (
	defaultDateFormatForFileName = "2006_01_02"
	defaultDateLayout            = "2006-01-02"
	defaultFullTimeLayout        = "2006-01-02 15:04:05.9999"
	defaultTimeLayout            = "2006-01-02 15:04:05"
)

func NewWriter(jiaLog *JiaLog) *writer {
	wr := &writer{
		writeChan: make(chan chanLog, 10000),
		jiaLog:    jiaLog,
	}
	go wr.handleWrite()
	return wr
}

func (w *writer) write(log string, logTarget string, logLevel string, isRaw bool) {

	skip := 3
	logCtx, err := callerInfo(skip)
	if err != nil {
		fmt.Println("log println err! " + time.Now().Format("2006-01-02 15:04:05") + " Error: " + err.Error())
		logCtx = &logContext{}
	}
	chanLog := chanLog{
		LogTarget: logTarget + "_" + logLevel,
		Content:   log + "\n",
		LogLevel:  logLevel,
		isRaw:     isRaw,
		logCtx:    logCtx,
	}

	w.writeChan <- chanLog
}

func (w *writer) handleWrite() {
	var chanLog chanLog
	var log string
	var logPath string

	for {
		chanLog = <-w.writeChan

		if !chanLog.isRaw {
			log = fmt.Sprintf(fmt.Sprintf("[%s] %s [%s:%v] %s", chanLog.LogLevel, time.Now().Format(defaultFullTimeLayout), chanLog.logCtx.fileName, chanLog.logCtx.line, chanLog.Content))
		} else {
			log = chanLog.Content
		}
		if w.jiaLog.enableConsole {
			fmt.Println(log)
		}
		if w.jiaLog.enableLog {
			logPath = w.jiaLog.logPath + chanLog.LogTarget
			w.writeFile(logPath+"_"+time.Now().Format(defaultDateFormatForFileName)+".log", log)
		}

	}
}

func (w *writer) writeFile(logPath string, content string) {
	pathDir := filepath.Dir(logPath)
	if !file.Exist(pathDir) {
		err := os.MkdirAll(pathDir, 0777)
		if err != nil {
			fmt.Println("logger.writeFile create path error ", err)
			return
		}
	}

	var mode os.FileMode
	flag := syscall.O_RDWR | syscall.O_APPEND | syscall.O_CREAT
	mode = 0666
	file, err := os.OpenFile(logPath, flag, mode)
	defer file.Close()
	if err != nil {
		fmt.Println(logPath, err)
		return
	}
	file.WriteString(content)
}
