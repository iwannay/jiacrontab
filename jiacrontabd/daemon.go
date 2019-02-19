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

		var (
			myCmdUint cmdUint
			stop      bool
			err       error
		)

		myCmdUint.ctx = ctx
		myCmdUint.dir = d.job.WorkDir
		myCmdUint.user = d.job.User
		myCmdUint.logName = fmt.Sprintf("%d.log", d.job.ID)
		myCmdUint.logPath = filepath.Join(cfg.LogPath, "daemon_job")

		log.Info("daemon exec job, jobName:", d.job.Name, " jobID", d.job.ID)

		err = myCmdUint.launch()

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

	if d.action == proto.ActionDeleteDaemonTask {
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
		var action int
		log.Info("init daemon task_name:", v.Name, "task_id:", v.ID, "status:", v.Status)
		job := v

		switch v.Status {
		case models.StatusJobOk:
			action = proto.ActionStartDaemonTask
		case models.StatusJobStop:
			action = proto.ActionStopDaemonTask
		default:
			continue
		}

		d.add(&daemonTask{
			job:    &job,
			action: action,
		})
	}

	go func() {
		var ctx context.Context
		for v := range d.taskChannel {
			switch v.action {
			case proto.ActionStartDaemonTask:
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
			case proto.ActionDeleteDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.job.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					model.DB().Unscoped().Delete(v.job, "id=?", v.job.ID)
					d.lock.Unlock()
				}
			case proto.ActionStopDaemonTask:
				d.lock.Lock()
				if t := d.taskMap[v.job.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					d.lock.Unlock()
					model.DB().Model(&model.DaemonTask{}).Where("id = ?", v.job.ID).Update("status", models.StatusJobStop)
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
