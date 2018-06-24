package main

import (
	"context"
	"fmt"
	"jiacrontab/model"
	"path/filepath"
	"sync"
	"time"
)

const (
	stopDaemonTask = iota
	startDaemonTask
	deleteDaemonTask
)

type daemonTask struct {
	task   *model.DaemonTask
	daemon *daemon

	action     int
	processNum int
}

func (d *daemonTask) do(ctx context.Context) {

	d.processNum = 1
	t := time.NewTimer(1 * time.Second)
	d.daemon.wait.Add(1)
	defer d.daemon.wait.Done()
	model.DB().Table("daemon_tasks").Table("daemon_tasks").Where("id = ?", d.task.ID).Update("status", startDaemonTask)

	for {
		var cmdList [][]string
		cmd := []string{d.task.Command, d.task.Args}
		stop := false
		cmdList = append(cmdList, cmd)
		var logContent []byte
		logPath := filepath.Join(globalConfig.logPath, "daemon_task")

		err := wrapExecScript(ctx, fmt.Sprintf("%d.log", d.task.ID), cmdList, logPath, &logContent)
		if err != nil {
			if d.task.MailNofity {
				sendMail(d.task.MailTo, globalConfig.addr+"提醒常驻脚本异常退出", fmt.Sprintf(
					"任务名：%s\n详情：%s %v\n开始时间：%s\n异常：%s",
					d.task.Name, d.task.Command, d.task.Args, time.Now().Format("2006-01-02 15:04:05"), err.Error()))
			}
		}

		select {
		case <-ctx.Done():
			stop = true
		case <-t.C:
		}

		if stop {
			t.Stop()
			break
		}

	}
	t.Stop()

	d.processNum = 0
	switch d.action {
	case deleteDaemonTask:

		d.daemon.lock.Lock()
		delete(d.daemon.taskMap, d.task.ID)
		d.daemon.lock.Unlock()

		model.DB().Table("daemon_tasks").Delete(d.task)
	case stopDaemonTask:

		d.daemon.lock.Lock()
		delete(d.daemon.taskMap, d.task.ID)
		d.daemon.lock.Unlock()

		model.DB().Table("daemon_tasks").Where("id = ?", d.task.ID).Update("status", stopDaemonTask)

	}

}

type daemon struct {
	taskChannel chan *daemonTask
	taskMap     map[uint]context.CancelFunc
	lock        sync.Mutex
	wait        sync.WaitGroup
}

func newDaemon(taskChannelLength int) *daemon {

	return &daemon{
		taskMap:     make(map[uint]context.CancelFunc),
		taskChannel: make(chan *daemonTask, taskChannelLength),
	}
}

func (d *daemon) add(t *daemonTask) {
	if t != nil {
		t.daemon = d
		d.taskChannel <- t
	}
}

func (d *daemon) run() {
	go func() {
		var ctx context.Context
		var cancel context.CancelFunc
		for v := range d.taskChannel {
			switch v.action {
			case startDaemonTask:
				d.lock.Lock()
				if d.taskMap[v.task.ID] == nil {
					ctx, cancel = context.WithCancel(context.Background())
					d.taskMap[v.task.ID] = cancel
					d.lock.Unlock()

					go v.do(ctx)

				} else {
					d.lock.Unlock()
				}
			case deleteDaemonTask:
				d.lock.Lock()
				if cancel = d.taskMap[v.task.ID]; cancel != nil {
					d.lock.Unlock()
					cancel()
				} else {
					d.lock.Unlock()
				}
			case stopDaemonTask:
				d.lock.Lock()
				if cancel = d.taskMap[v.task.ID]; cancel != nil {
					d.lock.Unlock()
					cancel()
				} else {
					d.lock.Unlock()
				}
			}

		}
	}()
}

func (d *daemon) waitDone() {
	d.wait.Wait()
}
