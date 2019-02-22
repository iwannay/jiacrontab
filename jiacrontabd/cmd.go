package jiacrontabd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"jiacrontab/pkg/kproc"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/util"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/iwannay/log"
)

type cmdUint struct {
	ctx     context.Context
	logName string
	args    [][]string
	logPath string
	content []byte
	logFile *os.File
	user    string
	env     []string
	dir     string
}

func (cu *cmdUint) release() {
	if cu.logFile != nil {
		cu.logFile.Close()
	}
}

func (cu *cmdUint) launch() error {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("wrapExecScript error:%v\n%s", err, debug.Stack())
		}
		cu.release()
	}()

	var err error

	if err = cu.setLogFile(); err != nil {
		return err
	}

	if len(cu.args) > 1 {
		err = cu.pipeExec()
	} else {
		err = cu.exec()
	}

	if err != nil {
		var errMsg string
		if cfg.VerboseJobLog {
			prefix := fmt.Sprintf("[%s %s %v] ", time.Now().Format(proto.DefaultTimeLayout), cfg.LocalAddr, cu.args)
			errMsg = prefix + err.Error() + "\n"
			cu.logFile.WriteString(errMsg)
		} else {
			errMsg = err.Error() + "\n"
			cu.logFile.WriteString(errMsg)
		}
		cu.content = append(cu.content, []byte(errMsg)...)

		return err
	}

	return nil
}

func (cu *cmdUint) setLogFile() error {
	var err error
	logPath := filepath.Join(cu.logPath, time.Now().Format("2006/01/02"))
	cu.logFile, err = util.TryOpen(filepath.Join(logPath, cu.logName), os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	return nil
}

func (cu *cmdUint) exec() error {
	log.Debug("args:", cu.args)
	cmdName := cu.args[0][0]
	args := cu.args[0][1:]
	cmd := kproc.CommandContext(cu.ctx, cmdName, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	defer stdout.Close()

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return err
	}

	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		return err
	}

	reader := bufio.NewReader(stdout)
	readerErr := bufio.NewReader(stderr)
	// 如果已经存在日志则直接写入
	cu.logFile.Write(cu.content)

	go func() {
		var (
			err  error
			line []byte
		)

		for {

			line, err = reader.ReadBytes('\n')
			if err != nil || err == io.EOF {
				break
			}

			if cfg.VerboseJobLog {
				prefix := fmt.Sprintf("[%s %s %s %v] ", time.Now().Format(proto.DefaultTimeLayout), cfg.LocalAddr, cmdName, args)
				line = append([]byte(prefix), line...)
				cu.content = append(cu.content, line...)
			} else {
				cu.content = append(cu.content, line...)
			}

			cu.logFile.Write(line)
		}

		for {
			line, err = readerErr.ReadBytes('\n')
			if err != nil || err == io.EOF {
				break
			}
			// 默认给err信息加上日期标志
			if cfg.VerboseJobLog {
				prefix := fmt.Sprintf("[%s %s %s %s] ", time.Now().Format(proto.DefaultTimeLayout), cfg.LocalAddr, cmdName, args)
				line = append([]byte(prefix), line...)
				cu.content = append(cu.content, line...)
			} else {
				cu.content = append(cu.content, line...)
			}
			cu.logFile.Write(line)
		}
	}()

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// ctx context.Context, cmdList [][]string, logname string, logpath string, content *[]byte
func (cu *cmdUint) pipeExec() error {
	var (
		outBufer       bytes.Buffer
		errBufer       bytes.Buffer
		cmdEntryList   []*pipeCmd
		err, exitError error
		line           []byte
	)

	for _, v := range cu.args {
		cmdName := v[0]
		args := v[1:]

		cmd := kproc.CommandContext(cu.ctx, cmdName, args...)
		cmdEntryList = append(cmdEntryList, &pipeCmd{cmd})
	}

	exitError = execute(&outBufer, &errBufer,
		cmdEntryList...,
	)

	// 如果已经存在日志则直接写入
	cu.logFile.Write(cu.content)

	for {

		line, err = outBufer.ReadBytes('\n')
		if err != nil || err == io.EOF {
			break
		}
		if cfg.VerboseJobLog {
			prefix := fmt.Sprintf("[%s %s %v] ", time.Now().Format(proto.DefaultTimeLayout), cfg.LocalAddr, cu.args)
			line = append([]byte(prefix), line...)
			cu.content = append(cu.content, line...)
		} else {
			cu.content = append(cu.content, line...)
		}

		cu.logFile.Write(line)
	}

	for {
		line, err = errBufer.ReadBytes('\n')
		if err != nil || err == io.EOF {
			break
		}
		// 默认给err信息加上日期标志
		if cfg.VerboseJobLog {
			prefix := fmt.Sprintf("[%s %s %v] ", time.Now().Format("2006-01-02 15:04:05"), cfg.LocalAddr, cu.args)
			line = append([]byte(prefix), line...)
			cu.content = append(cu.content, line...)
		} else {
			cu.content = append(cu.content, line...)
		}

		cu.logFile.Write(line)

	}

	return exitError

}

func writeLog(logpath string, logname string, content *[]byte) {
	logPath := filepath.Join(logpath, time.Now().Format("2006/01/02"))
	f, err := util.TryOpen(filepath.Join(logPath, logname), os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		log.Errorf("write log %v", err)
	}
	defer f.Close()
	f.Write(*content)
}

type pipeCmd struct {
	*kproc.KCmd
}

func execute(outputBuffer *bytes.Buffer, errorBuffer *bytes.Buffer, stack ...*pipeCmd) (err error) {
	pipeStack := make([]*io.PipeWriter, len(stack)-1)
	i := 0
	for ; i < len(stack)-1; i++ {
		stdinPipe, stdoutPipe := io.Pipe()
		stack[i].Stdout = stdoutPipe
		stack[i].Stderr = errorBuffer
		stack[i+1].Stdin = stdinPipe
		pipeStack[i] = stdoutPipe
	}

	stack[i].Stdout = outputBuffer
	stack[i].Stderr = errorBuffer

	if err = call(stack, pipeStack); err != nil {
		errorBuffer.WriteString(err.Error())
	}
	return err
}
