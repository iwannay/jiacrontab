package jiacrontabd

import (
	"bytes"
	"context"
	"fmt"
	"jiacrontab/pkg/proto"
	"path/filepath"
	"time"

	"github.com/iwannay/log"
)

type depEntry struct {
	jobID       uint   // 定时任务id
	jobUniqueID string // job的唯一标志
	processID   int    // 当前依赖的父级任务（可能存在多个并发的task)
	id          string // depID
	workDir     string
	user        string
	env         []string
	from        string
	commands    []string
	dest        string
	done        bool
	timeout     int64
	err         error
	name        string
	logContent  []byte
}

func newDependencies(crond *Jiacrontabd) *dependencies {
	return &dependencies{
		crond: crond,
		dep:   make(chan *depEntry, 100),
	}
}

type dependencies struct {
	crond *Jiacrontabd
	dep   chan *depEntry
}

func (d *dependencies) add(t *depEntry) {
	d.dep <- t
}

func (d *dependencies) run() {
	go func() {
		for {
			select {
			case t := <-d.dep:
				go d.exec(t)
			}
		}
	}()
}

func (d *dependencies) exec(task *depEntry) {

	var (
		reply     bool
		myCmdUnit cmdUint
		err       error
	)

	if task.timeout == 0 {
		// 默认超时10分钟
		task.timeout = 600
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.timeout)*time.Second)

	myCmdUnit.args = [][]string{task.commands}
	myCmdUnit.ctx = ctx
	myCmdUnit.dir = task.workDir
	myCmdUnit.user = task.user
	myCmdUnit.logName = fmt.Sprintf("%d-%s.log", task.jobID, task.id)
	myCmdUnit.logPath = filepath.Join(cfg.LogPath, "depend_job")

	err = myCmdUnit.launch()
	cancel()

	log.Infof("exec %s %s cost %.4fs %v", task.name, task.commands, float64(myCmdUnit.costTime)/1000000000, err)

	task.logContent = bytes.TrimRight(myCmdUnit.content, "\x00")
	task.done = true
	task.err = err

	task.dest, task.from = task.from, task.dest

	if !d.crond.SetDependDone(task) {
		err = rpcCall("Srv.SetDependDone", proto.DepJob{
			Name:        task.name,
			Dest:        task.dest,
			From:        task.from,
			ID:          task.id,
			JobUniqueID: task.jobUniqueID,
			ProcessID:   task.processID,
			JobID:       task.jobID,
			Commands:    task.commands,
			LogContent:  task.logContent,
			Err:         err,
			Timeout:     task.timeout,
		}, &reply)

		if err != nil {
			log.Error("Srv.SetDependDone error:", err, "server addr:", cfg.AdminAddr)
		}

		if !reply {
			log.Errorf("task %s %v call Srv.SetDependDone failed! err:%v", task.name, task.commands, err)
		}
	}
}
