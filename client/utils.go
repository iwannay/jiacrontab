package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func storge(data map[string]*proto.TaskArgs) error {
	var lock sync.RWMutex
	lock.Lock()

	f, err := libs.TryOpen(globalConfig.dataFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC)
	defer func() {
		f.Close()
		lock.Unlock()
	}()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(data, "", "  ")
	// b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

func checkMonth(check proto.CrontabArgs, month time.Month) bool {
	var flag = false
	if check.Month != "*" {
		if strings.Contains(check.Month, "/") {
			sli := strings.Split(check.Month, "/")
			tmp, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}
			if tmp == 0 {
				goto end
			}

			remainder := month % time.Month(tmp)
			if remainder == 0 {
				flag = true
			}

		} else if strings.Contains(check.Month, ",") {
			sli := strings.Split(check.Month, ",")
			for _, v := range sli {
				i, err := strconv.Atoi(v)
				if err != nil {
					log.Println(err)
					continue
				}
				if month.String() == proto.Months[i] {
					flag = true
					break
				}
			}
		} else if strings.Contains(check.Month, "-") {
			sli := strings.Split(check.Month, ",")
			lower, err := strconv.Atoi(sli[0])
			if err != nil {
				log.Println(err)
				goto end
			}
			upper, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}

			if month >= time.Month(lower) && month <= time.Month(upper) {
				flag = true
			}

		} else {
			m, err := strconv.Atoi(check.Month)
			if err != nil {
				goto end
			}
			if int(month) == m {
				flag = true
			}

		}

	} else {
		flag = true
	}
end:
	return flag
}

func checkWeekday(check proto.CrontabArgs, weekday time.Weekday) bool {
	var flag = false
	if check.Weekday != "*" {
		if strings.Contains(check.Weekday, "/") {
			sli := strings.Split(check.Weekday, "/")
			tmp, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}
			if tmp == 0 {
				goto end
			}

			remainder := weekday % time.Weekday(tmp)
			if remainder == 0 {
				flag = true
			}

		} else if strings.Contains(check.Weekday, ",") {
			sli := strings.Split(check.Weekday, ",")
			for _, v := range sli {
				i, err := strconv.Atoi(v)
				if err != nil {
					log.Println(err)
					continue
				}
				if weekday.String() == proto.Days[i] {
					flag = true
					break
				}
			}
		} else if strings.Contains(check.Weekday, "-") {
			sli := strings.Split(check.Weekday, "-")
			lower, err := strconv.Atoi(sli[0])
			if err != nil {
				log.Println(lower)
				goto end
			}
			upper, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(upper)
				goto end
			}
			if weekday >= time.Weekday(lower) && weekday <= time.Weekday(upper) {
				flag = true
			}
		} else {
			wd, err := strconv.Atoi(check.Weekday)
			if err != nil {
				goto end
			}
			if int(weekday) == wd {
				flag = true
			}

		}

	} else {
		flag = true
	}
end:
	return flag
}

func checkDay(check proto.CrontabArgs, day int) bool {
	var flag = false
	if check.Day != "*" {
		if strings.Contains(check.Day, "/") {
			sli := strings.Split(check.Day, "/")
			tmp, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}
			if tmp == 0 {
				goto end
			}

			remainder := (day - 1) % tmp
			if remainder == 0 {
				flag = true
			}

		} else if strings.Contains(check.Day, ",") {
			sli := strings.Split(check.Day, ",")
			for _, v := range sli {
				i, err := strconv.Atoi(v)
				if err != nil {
					log.Println(err)
					continue
				}
				if day == i {
					flag = true
					break
				}
			}
		} else if strings.Contains(check.Day, "-") {
			sli := strings.Split(check.Day, "-")
			lower, err := strconv.Atoi(sli[0])
			if err != nil {
				log.Println(err)
				goto end
			}
			upper, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}
			if day >= lower && day <= upper {
				flag = true
			}
		} else {
			i, err := strconv.Atoi(check.Day)
			if err != nil {
				log.Println(err)
				goto end
			}
			if i == day {
				flag = true
			}
		}

	} else {
		flag = true
	}
end:
	return flag
}

func checkHour(check proto.CrontabArgs, hour int) bool {
	var flag = false
	if check.Hour != "*" {
		if strings.Contains(check.Hour, "/") {
			sli := strings.Split(check.Hour, "/")
			tmp, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}
			if tmp == 0 {
				goto end
			}

			remainder := hour % tmp
			if remainder == 0 {
				flag = true
			}

		} else if strings.Contains(check.Hour, ",") {
			sli := strings.Split(check.Hour, ",")
			for _, v := range sli {
				i, err := strconv.Atoi(v)
				if err != nil {
					log.Println(err)
					continue
				}
				if hour == i {
					flag = true
					break
				}
			}
		} else if strings.Contains(check.Hour, "-") {
			sli := strings.Split(check.Hour, "-")
			lower, err := strconv.Atoi(sli[0])
			if err != nil {
				log.Println(err)
				goto end
			}
			upper, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}
			if hour >= lower && hour <= upper {
				flag = true
			}
		} else {
			i, err := strconv.Atoi(check.Hour)
			if err != nil {
				log.Println(err)
				goto end
			}
			if i == hour {
				flag = true
			}
		}

	} else {
		flag = true
	}
end:
	return flag
}

func checkMinute(check proto.CrontabArgs, minute int) bool {
	var flag = false
	if check.Minute != "*" {
		if strings.Contains(check.Minute, "/") {
			sli := strings.Split(check.Minute, "/")
			tmp, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}
			if tmp == 0 {
				goto end
			}
			remainder := minute % tmp
			if remainder == 0 {
				flag = true
			}

		} else if strings.Contains(check.Minute, ",") {
			sli := strings.Split(check.Minute, ",")
			for _, v := range sli {
				i, err := strconv.Atoi(v)
				if err != nil {
					log.Println(err)
					continue
				}
				if minute == i {
					flag = true
					break
				}
			}
		} else if strings.Contains(check.Minute, "-") {
			sli := strings.Split(check.Minute, "-")
			lower, err := strconv.Atoi(sli[0])
			if err != nil {
				log.Println(err)
				goto end
			}
			upper, err := strconv.Atoi(sli[1])
			if err != nil {
				log.Println(err)
				goto end
			}
			if minute >= lower && minute <= upper {
				flag = true
			}
		} else {
			i, err := strconv.Atoi(check.Minute)
			if err != nil {
				log.Println(err)
				goto end
			}
			if i == minute {
				flag = true
			}
		}

	} else {
		flag = true
	}
end:
	return flag
}

func wrapExecScript(ctx context.Context, logname string, cmdList [][]string, logpath string, content *[]byte) error {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
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
		args = strings.Split(cmdList[0][1], " ")
		f, err = execScript(ctx, logname, bin, logpath, content, args)
		cmdStr = bin + " " + strings.Join(args, " ")
	}

	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	if err != nil {
		var errMsg string
		if globalConfig.debugScript {
			prefix := fmt.Sprintf("[%s %s %s]>>  ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, cmdStr)
			errMsg = prefix + err.Error() + "\n"
			f.WriteString(errMsg)
		} else {
			errMsg = err.Error() + "\n"
			f.WriteString(errMsg)
		}
		*content = append(*content, []byte(errMsg)...)
	}

	return err
}

func execScript(ctx context.Context, logname string, bin string, logpath string, content *[]byte, args []string) (*os.File, error) {
	binpath, err := exec.LookPath(bin)
	if err != nil {
		return nil, err
	}

	logPath := filepath.Join(logpath, strconv.Itoa(time.Now().Year()), time.Now().Month().String())
	f, err := libs.TryOpen(filepath.Join(logPath, logname), os.O_APPEND|os.O_CREATE|os.O_RDWR)

	if err != nil {
		return f, err
	}

	cmd := exec.CommandContext(ctx, binpath, args...)
	stdout, err := cmd.StdoutPipe()
	defer stdout.Close()
	if err != nil {
		return f, err
	}
	stderr, err := cmd.StderrPipe()
	defer stderr.Close()
	if err != nil {
		return f, err
	}

	if err := cmd.Start(); err != nil {
		return f, err
	}
	reader := bufio.NewReader(stdout)
	readerErr := bufio.NewReader(stderr)
	// 如果已经存在日志则直接写入
	f.Write(*content)

	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if globalConfig.debugScript {
			prefix := fmt.Sprintf("[%s %s %s %s]>>  ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, bin, strings.Join(args, " "))
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
		if globalConfig.debugScript {
			prefix := fmt.Sprintf("[%s %s %s %s]>>  ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, bin, strings.Join(args, " "))
			line = prefix + line
			*content = append(*content, []byte(line)...)
		} else {
			*content = append(*content, []byte(line)...)
		}

		f.WriteString(line)

	}

	if err := cmd.Wait(); err != nil {
		return f, err
	}

	return f, nil
}

func pipeExecScript(ctx context.Context, cmdList [][]string, logname string, logpath string, content *[]byte) (*os.File, error) {
	var outBufer bytes.Buffer
	var errBufer bytes.Buffer
	var cmdEntryList []*exec.Cmd
	var f *os.File
	var logPath string
	var err, exitError error
	var logCmdName string

	for k, v := range cmdList {
		name := v[0]
		args := strings.Split(v[1], " ")
		if k > 0 {
			logCmdName += " | "
		}
		logCmdName += v[0] + " " + v[1]
		cmdEntryList = append(cmdEntryList, exec.CommandContext(ctx, name, args...))
	}

	exitError = execute(&outBufer, &errBufer,
		cmdEntryList...,
	)

	logPath = filepath.Join(logpath, strconv.Itoa(time.Now().Year()), time.Now().Month().String())
	f, err = libs.TryOpen(filepath.Join(logPath, logname), os.O_APPEND|os.O_CREATE|os.O_RDWR)
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
		if globalConfig.debugScript {
			prefix := fmt.Sprintf("[%s %s %s]>>  ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, logCmdName)
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
		if globalConfig.debugScript {
			prefix := fmt.Sprintf("[%s %s %s]>>  ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, logCmdName)
			line = prefix + line
			*content = append(*content, []byte(line)...)
		} else {
			*content = append(*content, []byte(line)...)
		}

		f.WriteString(line)

	}

	return f, exitError

}

func writeLog(logpath string, logname string, content *[]byte) {
	logPath := filepath.Join(logpath, strconv.Itoa(time.Now().Year()), time.Now().Month().String())
	f, err := libs.TryOpen(filepath.Join(logPath, logname), os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		log.Printf("write log %v", err)
	}
	defer f.Close()
	f.Write(*content)
}

func execute(outputBuffer *bytes.Buffer, errorBuffer *bytes.Buffer, stack ...*exec.Cmd) (err error) {
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

func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
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
			if err == nil {
				pipes[0].Close()
				err = call(stack[1:], pipes[1:])
			}
		}()
	}
	return stack[0].Wait()
}

func initPprof(addr string) {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	go func() {
		log.Printf("pprof listen %s", addr)
		if err := http.ListenAndServe(addr, pprofServeMux); err != nil {
			log.Printf("http.ListenAndServe(\"%s\", pprofServeMux) error(%v)", addr, err)
			panic(err)
		}
	}()

}

func sendMail(mailTo, title, content string) {
	if mailTo == "" {
		mailTo = globalConfig.mailTo
	}
	host := globalStore.Mail.Host
	from := globalStore.Mail.User
	pass := globalStore.Mail.Pass
	port := globalStore.Mail.Port

	go libs.SendMail(title, content, host, from, pass, port, mailTo)
}

func pushDepends(dpds []*dependScript) bool {

	if len(dpds) > 0 {
		var ndpds []proto.MScript
		for _, v := range dpds {
			// 检测目标服务器为本机时直接执行脚本
			if v.dest == globalConfig.addr {
				globalDepend.Add(v)
			} else {
				ndpds = append(ndpds, proto.MScript{
					Name:    v.name,
					Dest:    v.dest,
					From:    v.from,
					TaskId:  v.id,
					Command: v.command,
					Args:    v.args,
					Timeout: v.timeout,
				})
			}
		}
		if len(ndpds) > 0 {
			var reply bool
			err := rpcCall("Logic.Depends", ndpds, &reply)
			if !reply || err != nil {
				log.Printf("push Depends failed,%s", err)
				return false
			}
		}

	}
	return true
}

// 同步添加依赖执行
func pushPipeDepend(dpds []*dependScript, dependScriptId string) bool {
	var flag = true
	if len(dpds) > 0 {
		flag = false
		l := len(dpds) - 1
		for k, v := range dpds {
			if flag || dependScriptId == "" {

				// 检测目标服务器为本机时直接执行脚本
				log.Printf("sync push %s <%s %s>", v.dest, v.command, v.args)
				if v.dest == globalConfig.addr {
					globalDepend.Add(v)
				} else {
					var reply bool
					err := rpcCall("Logic.Depends", []proto.MScript{{
						Name:    v.name,
						Dest:    v.dest,
						From:    v.from,
						TaskId:  v.id,
						Command: v.command,
						Args:    v.args,
						Timeout: v.timeout,
					}}, &reply)
					if !reply || err != nil {
						log.Printf("sync push Depends failed!")
						return false
					}
				}
				flag = true
				break
			}

			if (v.id == dependScriptId) && (l != k) {
				flag = true
			}

		}

	}
	return flag
}

// filterDepend 本地执行的脚本依赖不再请求网络，直接转发到对应的处理模块
// 目标网络不是本机时返回false
func filterDepend(args *dependScript) bool {

	if args.dest != globalConfig.addr {
		return false
	}

	idArr := strings.Split(args.id, "-")
	isAllDone := true
	globalCrontab.lock.Lock()
	if h, ok := globalCrontab.handleMap[idArr[0]]; ok {

		globalCrontab.lock.Unlock()
		var logContent []byte
		var currTaskEntity *taskEntity
		for _, v := range h.taskPool {
			if v.id == idArr[1] {
				currTaskEntity = v
				for _, v2 := range v.depends {

					if v2.done == false {
						isAllDone = false
					} else {
						logContent = append(logContent, v2.logContent...)
					}

					if v2.id == args.id && v.sync {
						if ok := pushPipeDepend(v.depends, v2.id); ok {
							return true
						}
					}
				}
			}
		}

		if currTaskEntity == nil {
			log.Printf("cant find task entity %s %s %s", args.name, args.command, args.args)
			return true
		}

		// 如果依赖脚本执行出错直接通知主脚本停止
		if args.err != nil {
			isAllDone = true
			log.Printf("depend %s %s %s exec failed %s try to stop master task", args.name, args.command, args.args, args.err)
		}

		if isAllDone {
			currTaskEntity.ready <- struct{}{}
			currTaskEntity.logContent = logContent
		}

	} else {
		log.Printf("cant find task handle %s %s %s", args.name, args.command, args.args)
		globalCrontab.lock.Unlock()
	}

	return true

}
