package main

import (
	"context"
	"fmt"
	"jiacrontab/model"
	"log"
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
	task       *model.DaemonTask
	daemon     *daemon
	action     int
	cancel     context.CancelFunc
	processNum int
}

func (d *daemonTask) do(ctx context.Context) {

	d.processNum = 1
	t := time.NewTicker(1 * time.Second)
	d.daemon.wait.Add(1)
	defer d.daemon.wait.Done()
	model.DB().Table("daemon_tasks").Table("daemon_tasks").Where("id = ?", d.task.ID).Update(map[string]interface{}{
		"status":      startDaemonTask,
		"start_at":    time.Now(),
		"process_num": d.processNum})

	for {
		var cmdList [][]string
		cmd := []string{d.task.Command, d.task.Args}
		stop := false
		cmdList = append(cmdList, cmd)
		var logContent []byte
		logPath := filepath.Join(globalConfig.logPath, "daemon_task")
		log.Println("daemon exec task_name:", d.task.Name, "task_id", d.task.ID)
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
			break
		}

	}
	t.Stop()

	switch d.action {
	case deleteDaemonTask:

		d.daemon.lock.Lock()
		delete(d.daemon.taskMap, d.task.ID)
		d.daemon.lock.Unlock()

		model.DB().Delete(d.task, "id=?", d.task.ID)
	case stopDaemonTask:

		d.daemon.lock.Lock()
		delete(d.daemon.taskMap, d.task.ID)
		d.daemon.lock.Unlock()
	}

	d.processNum = 0
	model.DB().Table("daemon_tasks").Where("id = ?", d.task.ID).Update(map[string]interface{}{
		"status":      stopDaemonTask,
		"process_num": d.processNum})

	fmt.Println("end", d.task.Name)

}

type daemon struct {
	taskChannel chan *daemonTask
	taskMap     map[uint]*daemonTask
	lock        sync.Mutex
	wait        sync.WaitGroup
}

func newDaemon(taskChannelLength int) *daemon {

	return &daemon{
		taskMap:     make(map[uint]*daemonTask),
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

	// init daemon task
	var taskList []model.DaemonTask
	err := model.DB().Find(&taskList).Error
	if err != nil {
		log.Println("init daemon task error:", err)
	}
	for _, v := range taskList {
		log.Println("init daemon task_name:", v.Name, "task_id:", v.ID, "status:", v.Status)
		task := v
		d.add(&daemonTask{
			task:   &task,
			action: v.Status,
		})
	}

	go func() {
		var ctx context.Context

		for v := range d.taskChannel {

			switch v.action {
			case startDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.task.ID]; t == nil {
					d.taskMap[v.task.ID] = v
					d.lock.Unlock()
					ctx, v.cancel = context.WithCancel(context.Background())
					go v.do(ctx)
					log.Println("start", v.task.Name)
				} else {
					d.lock.Unlock()
					t.action = v.action
					if t.processNum == 0 {

						ctx, v.cancel = context.WithCancel(context.Background())
						go v.do(ctx)
					}

				}
			case deleteDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.task.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					model.DB().Delete(v.task, "id=?", v.task.ID)
					d.lock.Unlock()
				}
			case stopDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.task.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					d.lock.Unlock()
					model.DB().Table("daemon_tasks").Where("id = ?", v.task.ID).Update("status", stopDaemonTask)

				}
			}

		}
	}()
}

func (d *daemon) waitDone() {
	d.wait.Wait()
}
