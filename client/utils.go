package main

import (
	"bufio"
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

func wrapExecScript(ctx context.Context, logname string, bin string, logpath string, content *[]byte, args ...string) error {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	f, err := execScript(ctx, logname, bin, logpath, content, args...)
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	if err != nil {
		var errMsg string
		if globalConfig.debugScript {
			prefix := fmt.Sprintf("[%s %s %s %s]>>  ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, bin, strings.Join(args, " "))
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

func execScript(ctx context.Context, logname string, bin string, logpath string, content *[]byte, args ...string) (*os.File, error) {

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

func writeLog(logpath string, logname string, content *[]byte) {
	logPath := filepath.Join(logpath, strconv.Itoa(time.Now().Year()), time.Now().Month().String())
	f, err := libs.TryOpen(filepath.Join(logPath, logname), os.O_APPEND|os.O_CREATE|os.O_RDWR)
	if err != nil {
		log.Printf("write log %v", err)
	}
	defer f.Close()
	f.Write(*content)
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

func pushDepends(dpds []proto.MScript, curQueue proto.MScriptContent) bool {

	if len(dpds) > 0 {

		copyDpds := make([]proto.MScript, 0)
		copyDpds = append(copyDpds, dpds...)
		for k := range copyDpds {
			copyDpds[k].Queue = make([]proto.MScriptContent, 0)
			copyDpds[k].Queue = append(copyDpds[k].Queue, curQueue)
		}

		var ndpds []proto.MScript
		for _, v := range copyDpds {
			// 检测目标服务器为本机时直接执行脚本
			if v.Dest == globalConfig.addr {
				globalDepend.Add(v)
			} else {
				ndpds = append(ndpds, v)
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

// sign dest+cmd+args
// sign相同认为为同一个依赖,sign为空时取第一个
func pushPipeDepend(dpds []proto.MScript, sign string, curQueue proto.MScriptContent) bool {
	var flag = true
	if len(dpds) > 0 {
		var ndpds []proto.MScript
		flag = false

		// 解除引用
		copyDpds := make([]proto.MScript, 0)
		copyDpds = append(copyDpds, dpds...)
		for k := range copyDpds {
			copyDpds[k].Queue = make([]proto.MScriptContent, 0)
			copyDpds[k].Queue = append(copyDpds[k].Queue, curQueue)
		}

		flag = false
		l := len(copyDpds) - 1
		for k, v := range copyDpds {
			if flag || sign == "" {
				// 检测目标服务器为本机时直接执行脚本
				log.Printf("sync push %s <%s %s>", v.Dest, v.Command, v.Args)
				if v.Dest == globalConfig.addr {
					globalDepend.Add(v)
				} else {
					ndpds = append(ndpds, v)
					var reply bool
					err := rpcCall("Logic.Depends", ndpds, &reply)
					if !reply || err != nil {
						log.Printf("sync push Depends failed!")
						return false
					}
				}
				flag = true
				break
			}
			if (v.Dest+v.Command+v.Args == sign) && (l != k) {
				flag = true
			}

		}

	}
	return flag
}

// filterDepend 本地执行的脚本依赖不再请求网络，直接转发到对应的处理模块
// 目标网络不是本机时返回false
func filterDepend(args proto.MScript) bool {
	if args.Dest != globalConfig.addr {
		return false
	}

	if t, ok := globalStore.SearchTaskList(args.TaskId); ok {
		flag := true
		closeMsg := fmt.Sprintf("depend queue is close and stop wait depend %s %s done", args.Command, args.Args)
		if t.State != 2 {
			log.Println(closeMsg)
			return true
		}
		for k, v := range t.Depends {
			argsSign := args.From + args.Command + args.Args
			vSign := v.Dest + v.Command + v.Args
			if argsSign == vSign {
				if len(v.Queue) == 0 {
					log.Println(closeMsg, "queue length", 0)
					return true
				}
				globalCrontab.sliceLock.Lock()
				for key, value := range v.Queue {
					if value.TaskTime == args.Queue[0].TaskTime {
						if value.Done == true {
							globalCrontab.sliceLock.Unlock()
							log.Println(closeMsg, ",tasktime %d value.done eq true", value.TaskTime)
							return true
						}
						t.Depends[k].Queue[key] = args.Queue[0]
					}

				}
				globalCrontab.sliceLock.Unlock()

				if t.Sync {
					if ok := pushPipeDepend(t.Depends, argsSign, args.Queue[0]); ok {
						return true
					}
				}

			}

			// 判断是否所有依赖执行完毕
			for _, value := range v.Queue {
				if value.TaskTime == args.Queue[0].TaskTime {
					if value.Done == false {
						flag = false
					}
				}
			}

		}

		// 如果依赖脚本执行出错直接通知主脚本停止
		if args.Queue[0].Err != "" {
			flag = true
			log.Printf("task %s <%s %s> exec failed %s try to stop master task", args.TaskId, args.Args, args.Args, args.Queue[0].Err)
		}

		if flag {
			var logContent []byte
			for _, v := range t.Depends {
				for _, value := range v.Queue {
					if value.TaskTime == args.Queue[0].TaskTime {
						logContent = append(logContent, v.Queue[0].LogContent...)
					}
				}

			}
			globalCrontab.resolvedDepends(t, logContent, args.Queue[0].TaskTime, args.Queue[0].Err)
			log.Println("exec Task.ResolvedSDepends done")
		}
		return true
	}
	return false
}
