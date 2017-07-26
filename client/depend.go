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

					if t.Timeout == 0 {
						// 默认超时10分钟
						t.Timeout = 600
					}

					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.Timeout)*time.Second)
					args := strings.Split(t.Args, " ")
					name := filepath.Base(args[len(args)-1])
					startTime := time.Now()
					start := startTime.UnixNano()

					err := wrapExecScript(ctx, fmt.Sprintf("%s-%s.log", name, t.TaskId), t.Command, globalConfig.logPath, &logContent, args...)
					cancel()
					costTime := time.Now().UnixNano() - start
					log.Printf("exec task %s <%s %s> cost %.4fs %v", t.TaskId, t.Command, t.Args, float64(costTime)/1000000000, err)
					if err != nil {
						logContent = append(logContent, []byte(err.Error())...)
					}

					l := len(t.Queue)
					if l == 0 {
						log.Printf("task %s <%s %s> exec failed depend queue length %d ", t.TaskId, t.Command, t.Args, l)
						return
					}
					t.Queue[0].LogContent = bytes.TrimRight(logContent, "\x00")
					t.Queue[0].Done = true
					if err != nil {
						t.Queue[0].Err = err.Error()
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
