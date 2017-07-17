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
					ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
					args := strings.Split(t.Args, " ")
					name := filepath.Base(args[len(args)-1])
					start := time.Now().UnixNano()

					err := execScript(ctx, fmt.Sprintf("%s-%s.log", name, t.TaskId), t.Command, globalConfig.logPath, &t.LogContent, args...)
					cancel()
					costTime := time.Now().UnixNano() - start
					log.Printf("exec task %s <%s %s> cost %.4fs %v", t.TaskId, t.Command, t.Args, float64(costTime)/1000000000, err)
					if err != nil {
						t.LogContent = append(t.LogContent, []byte(err.Error())...)
					}
					t.LogContent = bytes.TrimRight(t.LogContent, "\x00")

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
