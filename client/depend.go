package main

import (
	"bytes"
	"context"
	"fmt"
	"jiacrontab/libs/proto"
	"log"
	"path/filepath"
	"strings"
	"time"
)

func newDepend() *depend {
	return &depend{
		depends: make(chan proto.MScript, 100),
	}
}

type depend struct {
	depends chan proto.MScript
}

func (d *depend) Add(t proto.MScript) {
	d.depends <- t
}

func (d *depend) run() {
	go func() {
		for {
			select {
			case t := <-d.depends:
				go func(t proto.MScript) {
					var reply bool
					var logContent []byte
					// var mailTo string
					// var task *proto.TaskArgs
					// var ok bool

					if t.Timeout == 0 {
						// 默认超时10分钟
						t.Timeout = 600
					}

					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.Timeout)*time.Second)
					args := strings.Split(t.Args, " ")
					name := filepath.Base(args[len(args)-1])
					startTime := time.Now()
					start := startTime.UnixNano()

					err := execScript(ctx, fmt.Sprintf("%s-%s.log", name, t.TaskId), t.Command, globalConfig.logPath, &logContent, args...)
					cancel()
					costTime := time.Now().UnixNano() - start
					log.Printf("exec task %s <%s %s> cost %.4fs %v", t.TaskId, t.Command, t.Args, float64(costTime)/1000000000, err)
					if err != nil {
						// t.LogContent = append(t.LogContent, []byte(err.Error())...)
						logContent = append(logContent, []byte(err.Error()+"\n")...)
						// if task, ok = globalStore.SearchTaskList(t.TaskId); ok {
						// 	mailTo = task.MailTo
						// } else {
						// 	mailTo = globalConfig.mailTo
						// }
						// sendMail(mailTo, globalConfig.addr+"提醒脚本异常退出", fmt.Sprintf(
						// 	"任务名：%s\n依赖：%s %v\n开始时间：%s\n异常：%ss",
						// 	t.TaskId, t.Command, t.Args, startTime.Format("2006-01-02 15:04:05"), err.Error()))
					}
					// t.LogContent = bytes.TrimRight(t.LogContent, "\x00")

					// 易得队列最后一个task即为该任务的时间标志
					l := len(t.Queue)
					if l == 0 {
						log.Printf("task %s <%s %s> exec failed depend queue length %d ", t.TaskId, t.Command, t.Args, l)
						return
					}
					t.Queue[l-1].LogContent = bytes.TrimRight(logContent, "\x00")
					t.Queue[l-1].Done = true
					if err != nil {
						t.Queue[l-1].Err = err.Error()
					}

					t.Dest, t.From = t.From, t.Dest

					if !filterDepend(t) {
						err = rpcCall("Logic.DependDone", t, &reply)
						if !reply || err != nil {
							log.Printf("task %s <%s %s> %s->Logic.DependDone failed!", t.TaskId, t.Command, t.Args, t.Dest)
						}
					}

				}(t)
			}

		}
	}()
}
