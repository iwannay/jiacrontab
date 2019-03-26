package jiacrontabd

import (
	"context"
	"encoding/json"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"path/filepath"
	"sync"
	"time"

	"github.com/iwannay/log"
)

type ApiNotifyArgs struct {
	JobName    string
	JobID      uint
	Commands   []string
	CreatedAt  time.Time
	NotifyType string
}

type daemonJob struct {
	job        *models.DaemonJob
	daemon     *Daemon
	action     int
	cancel     context.CancelFunc
	processNum int
}

func (d *daemonJob) do(ctx context.Context) {

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
			stop bool
			err  error
		)
		myCmdUint := cmdUint{
			ctx:     ctx,
			env:     d.job.WorkEnv,
			dir:     d.job.WorkDir,
			user:    d.job.WorkUser,
			logPath: filepath.Join(cfg.LogPath, "daemon_job", time.Now().Format("2006/01/02"), fmt.Sprintf("%d.log", d.job.ID)),
		}

		log.Info("daemon exec job, jobName:", d.job.Name, " jobID", d.job.ID)

		err = myCmdUint.launch()

		d.handleNotify(err)

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

	if d.action == proto.ActionDeleteDaemonJob {
		models.DB().Delete(d.job, "id=?", d.job.ID)
	}

	d.daemon.lock.Lock()
	delete(d.daemon.taskMap, d.job.ID)
	d.daemon.lock.Unlock()

	d.processNum = 0

	log.Info("daemon task end", d.job.Name)

}

func (d *daemonJob) handleNotify(err error) {
	if err == nil {
		return
	}

	var reply bool
	if d.job.ErrorMailNotify && len(d.job.MailTo) > 0 {
		var reply bool
		err := rpcCall("Srv.SendMail", proto.SendMail{
			MailTo:  d.job.MailTo,
			Subject: cfg.LocalAddr + "提醒常驻脚本异常退出",
			Content: fmt.Sprintf(
				"任务名：%s\n详情：%v\n开始时间：%s\n异常：%s",
				d.job.Name, d.job.Commands, time.Now().Format(proto.DefaultTimeLayout), err),
		}, &reply)
		if err != nil {
			log.Error("Srv.SendMail error:", err, "server addr:", cfg.AdminAddr)
		}
	}

	if d.job.ErrorAPINotify && len(d.job.APITo) > 0 {
		postData, err := json.Marshal(ApiNotifyArgs{
			JobName:    d.job.Name,
			JobID:      d.job.ID,
			Commands:   d.job.Commands,
			CreatedAt:  d.job.CreatedAt,
			NotifyType: "error",
		})
		if err != nil {
			log.Error("json.Marshal error:", err)
		}
		err = rpcCall("Srv.ApiPost", proto.ApiPost{
			Urls: d.job.APITo,
			Data: string(postData),
		}, &reply)

		if err != nil {
			log.Error("Logic.ApiPost error:", err, "server addr:", cfg.AdminAddr)
		}
	}
}

type Daemon struct {
	taskChannel chan *daemonJob
	taskMap     map[uint]*daemonJob
	lock        sync.Mutex
	wait        sync.WaitGroup
}

func newDaemon(taskChannelLength int) *Daemon {
	return &Daemon{
		taskMap:     make(map[uint]*daemonJob),
		taskChannel: make(chan *daemonJob, taskChannelLength),
	}
}

func (d *Daemon) add(t *daemonJob) {
	if t != nil {
		log.Infof("daemon.add(%s)\n", t.job.Name)
		t.daemon = d
		d.taskChannel <- t
	}
}

func (d *Daemon) run() {

	var jobList []models.DaemonJob
	err := models.DB().Find(&jobList).Error
	if err != nil {
		log.Error("init daemon task error:", err)
	}

	for _, v := range jobList {
		var action int
		log.Info("init daemon task_name:", v.Name, "task_id:", v.ID, "status:", v.Status)
		job := v

		switch v.Status {
		case models.StatusJobOk:
			action = proto.ActionStartDaemonJob
		case models.StatusJobStop:
			action = proto.ActionStopDaemonJob
		default:
			continue
		}

		d.add(&daemonJob{
			job:    &job,
			action: action,
		})
	}

	go func() {
		var ctx context.Context
		for v := range d.taskChannel {
			switch v.action {
			case proto.ActionStartDaemonJob:
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
			case proto.ActionDeleteDaemonJob:
				d.lock.Lock()
				if t := d.taskMap[v.job.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					models.DB().Delete(v.job, "id=?", v.job.ID)
					d.lock.Unlock()
				}
			case proto.ActionStopDaemonJob:
				d.lock.Lock()
				if t := d.taskMap[v.job.ID]; t != nil {
					d.lock.Unlock()
					t.action = v.action
					t.cancel()
				} else {
					d.lock.Unlock()
					models.DB().Model(&models.DaemonJob{}).Where("id = ?", v.job.ID).Update("status", models.StatusJobStop)
				}
			}
		}
	}()
}

func (d *Daemon) count() int {
	var count int
	d.lock.Lock()
	count = len(d.taskMap)
	d.lock.Unlock()
	return count
}

func (d *Daemon) waitDone() {
	d.wait.Wait()
}
