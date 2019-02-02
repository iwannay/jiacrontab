package jiacrontabd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"jiacrontab/pkg/kproc"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/rpc"
	"jiacrontab/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func wrapExecScript(ctx context.Context, logname string, cmdList [][]string, logpath string, content *[]byte) error {

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("wrapExecScript error:%v", err)
		}
	}()

	var err error
	var f *os.File
	var cmdStr string
	var bin string
	var args []string
	if len(cmdList) > 1 {
		for k, v := range cmdList {
			if k > 0 {
				cmdStr += " | "
			}
			cmdStr += v[0] + "  " + v[1]
		}
		f, err = pipeExecScript(ctx, cmdList, logname, logpath, content)
	} else {
		bin = cmdList[0][0]
		rawArgs := strings.Split(cmdList[0][1], " ")
		for _, v := range rawArgs {
			if strings.TrimSpace(v) != "" {
				args = append(args, v)
			}
		}

		f, err = execScript(ctx, logname, bin, logpath, content, args)
		cmdStr = bin + " " + strings.Join(args, " ")
	}

	if f != nil {
		defer f.Close()
	}

	if err != nil {
		var errMsg string
		if cfg.VerboseJobLog && f != nil {
			prefix := fmt.Sprintf("[%s %s %s] ", time.Now().Format("2006-01-02 15:04:05"), cfg.LocalAddr, cmdStr)
			errMsg = prefix + err.Error() + "\n"
			f.WriteString(errMsg)
		} else if f != nil {
			errMsg = err.Error() + "\n"
			f.WriteString(errMsg)
		}
		*content = append(*content, []byte(errMsg)...)

		return err
	}

	return err
}

func execScript(ctx context.Context, logname string, bin string, logpath string, content *[]byte, args []string) (*os.File, error) {

	logPath := filepath.Join(logpath, time.Now().Format("2006/01/02"))
	f, err := util.TryOpen(filepath.Join(logPath, logname), os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		return f, err
	}

	binpath, err := exec.LookPath(bin)
	if err != nil {
		return f, err
	}

	cmd := kproc.CommandContext(ctx, binpath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return f, err
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()

	if err != nil {
		return f, err
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		return f, err
	}

	reader := bufio.NewReader(stdout)
	readerErr := bufio.NewReader(stderr)
	// 如果已经存在日志则直接写入
	f.Write(*content)

	go func() {
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			if cfg.VerboseJobLog {
				prefix := fmt.Sprintf("[%s %s %s %s] ", time.Now().Format("2006-01-02 15:04:05"), cfg.LocalAddr, bin, strings.Join(args, " "))
				line = prefix + line
				*content = append(*content, []byte(line)...)
			} else {
				*content = append(*content, []byte(line)...)
			}

			f.WriteString(line)
		}

		for {
			line, err2 := readerErr.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			// 默认给err信息加上日期标志
			if cfg.VerboseJobLog {
				prefix := fmt.Sprintf("[%s %s %s %s] ", time.Now().Format("2006-01-02 15:04:05"), cfg.LocalAddr, bin, strings.Join(args, " "))
				line = prefix + line
				*content = append(*content, []byte(line)...)
			} else {
				*content = append(*content, []byte(line)...)
			}

			f.WriteString(line)
		}
	}()

	if err := cmd.Wait(); err != nil {
		return f, err
	}

	return f, nil
}

func pipeExecScript(ctx context.Context, cmdList [][]string, logname string, logpath string, content *[]byte) (*os.File, error) {
	var outBufer bytes.Buffer
	var errBufer bytes.Buffer
	var cmdEntryList []*pipeCmd
	var f *os.File
	var logPath string
	var err, exitError error
	var logCmdName string

	for k, v := range cmdList {
		name := v[0]
		rawArgs := strings.Split(v[1], " ")
		args := []string{}
		for _, v := range rawArgs {
			if strings.TrimSpace(v) != "" {
				args = append(args, v)
			}
		}

		if k > 0 {
			logCmdName += " | "
		}
		logCmdName += v[0] + " " + v[1]
		cmd := kproc.CommandContext(ctx, name, args...)
		cmdEntryList = append(cmdEntryList, &pipeCmd{cmd})
	}

	exitError = execute(&outBufer, &errBufer,
		cmdEntryList...,
	)

	logPath = filepath.Join(logpath, time.Now().Format("2006/01/02"))
	f, err = util.TryOpen(filepath.Join(logPath, logname), os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		return f, err
	}

	// 如果已经存在日志则直接写入
	f.Write(*content)

	for {
		line, err2 := outBufer.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if cfg.VerboseJobLog {
			prefix := fmt.Sprintf("[%s %s %s] ", time.Now().Format("2006-01-02 15:04:05"), cfg.LocalAddr, logCmdName)
			line = prefix + line
			*content = append(*content, []byte(line)...)
		} else {
			*content = append(*content, []byte(line)...)
		}

		f.WriteString(line)

	}

	for {
		line, err2 := errBufer.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		// 默认给err信息加上日期标志
		if cfg.VerboseJobLog {
			prefix := fmt.Sprintf("[%s %s %s] ", time.Now().Format("2006-01-02 15:04:05"), cfg.LocalAddr, logCmdName)
			line = prefix + line
			*content = append(*content, []byte(line)...)
		} else {
			*content = append(*content, []byte(line)...)
		}

		f.WriteString(line)

	}

	return f, exitError

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

func rpcCall(serviceMethod string, args, reply interface{}) error {
	return rpc.Call(cfg.AdminAddr, serviceMethod, args, reply)
}
