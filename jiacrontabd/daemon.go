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
	job        *models.DaemonJob
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
			log.Errorf("%s exec panic %s \n", d.job.Name, err)
		}

		d.daemon.wait.Done()
	}()

	for {
		var cmdList [][]string
		var logContent []byte

		stop := false
		cmdList = append(cmdList, d.job.Commands)

		logPath := filepath.Join(cfg.LogPath, "daemon_job")
		log.Info("daemon exec jobName:", d.job.Name, " jobID", d.job.ID)
		err := wrapExecScript(ctx, fmt.Sprintf("%d.log", d.job.ID), cmdList, logPath, &logContent)
		if err != nil {
			if d.job.ErrorMailNotify && d.job.MailTo != "" {
				err := rpcCall("Srv.SendMail", proto.SendMail{
					MailTo:  strings.Split(d.job.MailTo, ","),
					Subject: cfg.LocalAddr + "提醒常驻脚本异常退出",
					Content: fmt.Sprintf(
						"任务名：%s\n详情：%v\n开始时间：%s\n异常：%s", d.job.Name, d.job.Commands, time.Now().Format("2006-01-02 15:04:05"), err.Error()),
				}, &reply)
				if err != nil {
					log.Error("Logic.SendMail error:", err, "server addr:", cfg.AdminAddr)
				}
			}

			if d.job.ErrorAPINotify && d.job.APITo != "" {
				postData, err := json.Marshal(apiPost{
					JobName:   d.job.Name,
					JobID:     d.job.ID,
					Commmands: d.job.Commands,
					CreatedAt: d.job.CreatedAt,
					Type:      "error",
				})
				if err != nil {
					log.Error("json.Marshal error:", err)
				}
				err = rpcCall("Srv.ApiPost", proto.ApiPost{
					Url:  d.job.APITo,
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

		if stop || d.job.FailRestart == false {
			break
		}

	}
	t.Stop()

	if d.action == proto.DeleteDaemonTask {
		// model.DB().Unscoped().Delete(d.task, "id=?", d.task.ID)
	}

	d.daemon.lock.Lock()
	delete(d.daemon.taskMap, d.job.ID)
	d.daemon.lock.Unlock()

	d.processNum = 0

	log.Info("daemon task end", d.job.Name)

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
		log.Infof("daemon.add(%s)\n", t.job.Name)
		t.daemon = d
		d.taskChannel <- t
	}
}

func (d *daemon) run() {

	var jobList []models.DaemonJob
	err := model.DB().Find(&jobList).Error
	if err != nil {
		log.Error("init daemon task error:", err)
	}

	for _, v := range jobList {
		log.Info("init daemon task_name:", v.Name, "task_id:", v.ID, "status:", v.Status)
		job := v
		d.add(&daemonTask{
			job:    &job,
			action: v.Status,
		})
	}

	go func() {
		var ctx context.Context
		for v := range d.taskChannel {
			switch v.action {
			case startDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.job.ID]; t == nil {
					d.taskMap[v.job.ID] = v
					d.lock.Unlock()
					ctx, v.cancel = context.WithCancel(context.Background())
					go v.do(ctx)
					log.Info("start", v.job.Name)
				} else {
					d.lock.Unlock()

				}
			case deleteDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.job.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					model.DB().Unscoped().Delete(v.job, "id=?", v.job.ID)
					d.lock.Unlock()
				}
			case stopDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.job.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					d.lock.Unlock()
					model.DB().Model(&model.DaemonTask{}).Where("id = ?", v.job.ID).Update("status", stopDaemonTask)
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
