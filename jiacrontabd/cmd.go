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
        "strings"
        "strconv"
	"runtime/debug"
	"time"

	"github.com/iwannay/log"
)

type cmdUint struct {
	ctx              context.Context
	args             [][]string
	logPath          string
	content          []byte
	logFile          *os.File
	label            string
	user             string
	verboseLog       bool
	exportLog        bool
	env              []string
	killChildProcess bool
	dir              string
	startTime        time.Time
	costTime         time.Duration
	jd               *Jiacrontabd
}

func (cu *cmdUint) release() {
	if cu.logFile != nil {
		cu.logFile.Close()
	}
	cu.costTime = time.Now().Sub(cu.startTime)
}

func (cu *cmdUint) launch() error {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("wrapExecScript error:%v\n%s", err, debug.Stack())
		}
		cu.release()
	}()
	cfg := cu.jd.getOpts()
	cu.startTime = time.Now()

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
		if cu.verboseLog {
			prefix := fmt.Sprintf("[%s %s %s] ", time.Now().Format(proto.DefaultTimeLayout), cfg.BoardcastAddr, cu.label)
			errMsg = prefix + err.Error() + "\n"
		} else {
			errMsg = err.Error() + "\n"
		}

		cu.logFile.WriteString(errMsg)

		if cu.exportLog {
			cu.content = append(cu.content, []byte(errMsg)...)
		}

		return err
	}

	return nil
}

func (cu *cmdUint) setLogFile() error {
	var err error
	cu.logFile, err = util.TryOpen(cu.logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	return nil
}

func (cu *cmdUint) exec() error {
	log.Debug("cmd exec args:", cu.args)
	
        cmdName := cu.args[0][0]
	args := cu.args[0][1:]
	cmd := kproc.CommandContext(cu.ctx, cmdName, args...)
	cfg := cu.jd.getOpts()

	cmd.SetDir(cu.dir)
	cmd.SetEnv(cu.env)
	cmd.SetUser(cu.user)
	cmd.SetExitKillChildProcess(cu.killChildProcess)

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
        for _, args := range(cu.args) {
          for _, arg := range(args) {
            cu.logFile.WriteString(arg + "\n")
          }
          cu.logFile.WriteString(strconv.Itoa(len(args)) + " : " + strings.Join(args, " ") + "\n")
        }

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
				prefix := fmt.Sprintf("[%s %s %s] ", time.Now().Format(proto.DefaultTimeLayout), cfg.BoardcastAddr, cu.label)
				line = append([]byte(prefix), line...)
			}

			if cu.exportLog {
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
				prefix := fmt.Sprintf("[%s %s %s] ", time.Now().Format(proto.DefaultTimeLayout), cfg.BoardcastAddr, cu.label)
				line = append([]byte(prefix), line...)
			}
			if cu.exportLog {
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

func (cu *cmdUint) pipeExec() error {
	var (
		outBufer       bytes.Buffer
		errBufer       bytes.Buffer
		cmdEntryList   []*pipeCmd
		err, exitError error
		line           []byte
		cfg            = cu.jd.getOpts()
	)

	for _, v := range cu.args {
		cmdName := v[0]
		args := v[1:]

		cmd := kproc.CommandContext(cu.ctx, cmdName, args...)

		cmd.SetDir(cu.dir)
		cmd.SetEnv(cu.env)
		cmd.SetUser(cu.user)
		cmd.SetExitKillChildProcess(cu.killChildProcess)

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
			prefix := fmt.Sprintf("[%s %s %s] ", time.Now().Format(proto.DefaultTimeLayout), cfg.BoardcastAddr, cu.label)
			line = append([]byte(prefix), line...)
		}

		cu.content = append(cu.content, line...)
		cu.logFile.Write(line)
	}

	for {
		line, err = errBufer.ReadBytes('\n')
		if err != nil || err == io.EOF {
			break
		}

		if cfg.VerboseJobLog {
			prefix := fmt.Sprintf("[%s %s %s] ", time.Now().Format(proto.DefaultTimeLayout), cfg.BoardcastAddr, cu.label)
			line = append([]byte(prefix), line...)
		}

		if cu.exportLog {
			cu.content = append(cu.content, line...)
		}
		cu.logFile.Write(line)
	}
	return exitError
}

func call(stack []*pipeCmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}

	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}

		defer func() {
			pipes[0].Close()
			if err == nil {
				err = call(stack[1:], pipes[1:])
			}
			if err != nil {
				// fixed zombie process
				stack[1].Wait()
			}
		}()
	}
	return stack[0].Wait()
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
