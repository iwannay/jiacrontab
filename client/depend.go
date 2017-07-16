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
				var reply bool
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
				args := strings.Split(t.Args, " ")
				name := filepath.Base(args[len(args)-1])
				log.Printf("task %s exec depend %s %s", t.TaskId, t.Command, t.Args)
				err := execScript(ctx, fmt.Sprintf("%s-%s.log", name, t.TaskId), t.Command, globalConfig.logPath, &t.LogContent, args...)
				cancel()

				if err != nil {
					t.LogContent = append(t.LogContent, []byte(err.Error())...)
				}
				t.LogContent = bytes.TrimRight(t.LogContent, "\x00")

				t.Dest, t.From = t.From, t.Dest
				log.Println("rpcCall", "Logic.DependDone", t, &reply)
				err = rpcCall("Logic.DependDone", t, &reply)
				if !reply || err != nil {
					log.Printf("push %v failed!", t)
				}
			}

		}
	}()
}
