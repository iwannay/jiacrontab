package jiacrontabd

import (
	"context"
	"encoding/json"
	"fmt"
	"jiacrontab/model"
	"jiacrontab/models"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/proto"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	stopDaemonTask = iota
	startDaemonTask
	deleteDaemonTask
)

type daemonTask struct {
	task       *models.DaemonJob
	daemon     *daemon
	action     int
	cancel     context.CancelFunc
	processNum int
}

func (d *daemonTask) do(ctx context.Context) {
	type apiPost struct {
		JobName   string
		JobID     uint
		Commmands []string
		CreatedAt time.Time
		Type      string
	}

	var reply bool
	d.processNum = 1
	t := time.NewTicker(1 * time.Second)
	d.daemon.wait.Add(1)

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s exec panic %s \n", d.task.Name, err)
		}

		d.daemon.wait.Done()
	}()

	// model.DB().Table("daemon_tasks").Table("daemon_tasks").Where("id = ?", d.task.ID).Update(map[string]interface{}{
	// 	"status":      startDaemonTask,
	// 	"start_at":    time.Now(),
	// 	"process_num": d.processNum})

	for {
		var cmdList [][]string
		var logContent []byte

		stop := false
		cmdList = append(cmdList, d.task.Commands)

		logPath := filepath.Join(cfg.LogPath, "daemon_task")
		log.Info("daemon exec task_name:", d.task.Name, "task_id", d.task.ID)
		err := wrapExecScript(ctx, fmt.Sprintf("%d.log", d.task.ID), cmdList, logPath, &logContent)
		if err != nil {
			if d.task.MailNotify && d.task.MailTo != "" {

				err := rpcCall("Logic.SendMail", proto.SendMail{
					MailTo:  strings.Split(d.task.MailTo, ","),
					Subject: cfg.LocalAddr + "提醒常驻脚本异常退出",
					Content: fmt.Sprintf(
						"任务名：%s\n详情：%v\n开始时间：%s\n异常：%s", d.task.Name, d.task.Commands, time.Now().Format("2006-01-02 15:04:05"), err.Error()),
				}, &reply)
				if err != nil {
					log.Error("Logic.SendMail error:", err, "server addr:", cfg.AdminAddr)
				}
			}

			if d.task.ApiNotify && d.task.ApiTo != "" {
				postData, err := json.Marshal(apiPost{
					JobName:   d.task.Name,
					JobID:     d.task.ID,
					Commmands: d.task.Commands,
					CreatedAt: d.task.CreatedAt,
					Type:      "error",
				})
				if err != nil {
					log.Error("json.Marshal error:", err)
				}
				err = rpcCall("Logic.ApiPost", proto.ApiPost{
					Url:  d.task.ApiTo,
					Data: string(postData),
				}, &reply)
				if err != nil {
					log.Error("Logic.ApiPost error:", err, "server addr:", cfg.AdminAddr)
				}
			}
		}

		select {
		case <-ctx.Done():
			stop = true
		case <-t.C:
		}

		if stop || d.task.FailedRestart == false {
			break
		}

	}
	t.Stop()

	if d.action == proto.DeleteDaemonTask {
		// model.DB().Unscoped().Delete(d.task, "id=?", d.task.ID)
	}

	d.daemon.lock.Lock()
	delete(d.daemon.taskMap, d.task.ID)
	d.daemon.lock.Unlock()

	d.processNum = 0
	// model.DB().Model(&model.DaemonTask{}).Where("id = ?", d.task.ID).Update(map[string]interface{}{
	// 	"status":      stopDaemonTask,
	// 	"process_num": d.processNum})

	log.Info("daemon task end", d.task.Name)

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
		log.Info("daemon.add(%s)\n", t.task.Name)
		t.daemon = d
		d.taskChannel <- t
	}
}

func (d *daemon) run() {

	// init daemon task
	var taskList []models.DaemonJob
	err := model.DB().Find(&taskList).Error
	if err != nil {
		log.Error("init daemon task error:", err)
	}
	for _, v := range taskList {
		log.Info("init daemon task_name:", v.Name, "task_id:", v.ID, "status:", v.Status)
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
					log.Info("start", v.task.Name)
				} else {
					d.lock.Unlock()

				}
			case deleteDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.task.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					model.DB().Unscoped().Delete(v.task, "id=?", v.task.ID)
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
					model.DB().Model(&model.DaemonTask{}).Where("id = ?", v.task.ID).Update("status", stopDaemonTask)

				}
			}

		}
	}()
}

func (d *daemon) count() int {
	var count int
	d.lock.Lock()
	count = len(d.taskMap)
	d.lock.Unlock()
	return count
}

func (d *daemon) waitDone() {
	d.wait.Wait()
}
