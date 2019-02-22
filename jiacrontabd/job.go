package jiacrontabd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/crontab"
	"github.com/iwannay/log"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/util"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	exitError       = "error"
	exitKilled      = "kill"
	exitSuccess     = "success"
	exitDependError = "depend error"
	exitTimeout     = "timeout"
)

type process struct {
	id        int
	deps      []*depEntry
	ctx       context.Context
	cancel    context.CancelFunc
	logPath   string
	startTime time.Time
	endTime   time.Time
	jobEntry  *JobEntry
}

func newProcess(id int, jobEntry *JobEntry) *process {
	p := &process{
		id:       id,
		jobEntry: jobEntry,
		logPath:  filepath.Join(cfg.LogPath, "crontab_task"),
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())

	for _, v := range p.jobEntry.detail.DependJobs {
		saveID := fmt.Sprintf("dep-%d-%d-%s", p.jobEntry.detail.ID, p.id, v.ID)
		p.deps = append(p.deps, &depEntry{
			jobID:     p.jobEntry.detail.ID,
			processID: id,
			saveID:    saveID,
			from:      v.From,
			commands:  v.Commands,
			dest:      v.Dest,
			done:      false,
			timeout:   v.Timeout,
		})
	}

	return p
}

func (p *process) waitDepExecDone() bool {

	if len(p.deps) == 0 {
		log.Infof("%s jobID:%d has no dep", p.jobEntry.detail.Name, p.jobEntry.detail.ID)
		return true
	}

	syncFlag := true
	if p.jobEntry.detail.IsSync {
		// 同步
		syncFlag = p.jobEntry.jd.pushPipeDepend(p.deps, "")
	} else {
		// 并发模式
		syncFlag = p.jobEntry.jd.pushDepend(p.deps)
	}

	if !syncFlag {
		prefix := fmt.Sprintf("[%s %s] ", time.Now().Format("2006-01-02 15:04:05"), cfg.LocalAddr)
		p.jobEntry.logContent = append(p.jobEntry.logContent, []byte(prefix+"failed to exec depends, push depends error\n")...)
		return syncFlag
	}

	c := time.NewTimer(3600 * time.Second)
	defer c.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return false
		case <-c.C:
			log.Errorf("jobID:%d exec dep timeout!", p.jobEntry.detail.ID)
			return false
		case <-p.jobEntry.ready:
			log.Infof("jobID:%d exec all dep done.", p.jobEntry.detail.ID)
		}
	}
}

func (p *process) exec(logContent *[]byte) {
	var (
		reply     bool
		isTimeout bool
		err       error
		done      bool
		myCmdUint cmdUint
	)

	type errAPIPost struct {
		JobName   string
		JobID     int
		Commands  [][]string
		CreatedAt time.Time
		Timeout   int64
		Type      string
	}

	p.startTime = time.Now()
	if ok := p.waitDepExecDone(); !ok {
		errMsg := fmt.Sprintf("[%s %s %s] Execution of dependency job failed\n", time.Now().Format(proto.DefaultTimeLayout), cfg.LocalAddr, p.jobEntry.detail.Name)
		p.jobEntry.logContent = append(p.jobEntry.logContent, []byte(errMsg)...)
		writeLog(p.logPath, fmt.Sprintf("%d.log", p.jobEntry.detail.ID), &p.jobEntry.logContent)
		p.jobEntry.detail.LastExitStatus = exitDependError
		if p.jobEntry.detail.ErrorMailNotify && len(p.jobEntry.detail.MailTo) != 0 {
			p.endTime = time.Now()
			if err := rpcCall("Srv.SendMail", proto.SendMail{
				MailTo:  p.jobEntry.detail.MailTo,
				Subject: cfg.LocalAddr + "提醒脚本依赖异常退出",
				Content: fmt.Sprintf(
					"任务名：%s\n详情：%v\n开始时间：%s\n耗时：%.4f\n异常：%s",
					p.jobEntry.detail.Name, p.jobEntry.detail.Commands, p.endTime.Format(proto.DefaultTimeLayout), p.endTime.Sub(p.startTime).Seconds(), errors.New(errMsg)),
			}, &reply); err != nil {
				log.Error("Srv.SendMail error:", err, "server addr:", cfg.AdminAddr)
			}
		}
	} else {

		if p.jobEntry.detail.Timeout != 0 {
			time.AfterFunc(
				time.Duration(p.jobEntry.detail.Timeout)*time.Second, func() {
					if done {
						isTimeout = true
						switch p.jobEntry.detail.TimeoutTrigger {
						case "api":
							p.jobEntry.detail.LastExitStatus = exitTimeout
							postData, err := json.Marshal(errAPIPost{
								JobName:   p.jobEntry.detail.Name,
								JobID:     int(p.jobEntry.detail.ID),
								Commands:  p.jobEntry.detail.Commands,
								CreatedAt: p.jobEntry.detail.CreatedAt,
								Timeout:   int64(p.jobEntry.detail.Timeout),
								Type:      "timeout",
							})
							if err != nil {
								log.Error("json.Marshal error:", err)
								return
							}
							for _, url := range p.jobEntry.detail.APITo {
								if err = rpcCall("Srv.ErrorNotify err:", proto.ApiPost{
									Url:  url,
									Data: string(postData),
								}, &reply); err != nil {
									log.Error("Srv.ErrorNotify err:", err, "server addr:", cfg.AdminAddr)
								}
							}

						case "email":
							p.jobEntry.detail.LastExitStatus = exitTimeout
							if err = rpcCall("Srv.SendMail", proto.SendMail{
								MailTo:  p.jobEntry.detail.MailTo,
								Subject: cfg.LocalAddr + "提醒脚本执行超时",
								Content: fmt.Sprintf(
									"任务名：%s\n详情：%v\n开始时间：%s\n超时：%ds",
									p.jobEntry.detail.Name, p.jobEntry.detail.Commands, p.endTime.Format(proto.DefaultTimeLayout), p.jobEntry.detail.Timeout),
							}, &reply); err != nil {
								log.Error("Srv.SendMail error:", err, "server addr:", cfg.AdminAddr)
							}
						case "kill":
							p.jobEntry.detail.LastExitStatus = exitTimeout
							p.cancel()
						case "email_and_kill":
							p.jobEntry.detail.LastExitStatus = exitTimeout
							p.cancel()
							if err = rpcCall("Srv.SendMail", proto.SendMail{}, &reply); err != nil {
								log.Error("Srv.SendMail error:", err, "server addr:", cfg.AdminAddr)
							}
						case "ignore":
							isTimeout = false
						default:
						}
					}

				})
		}

		myCmdUint.ctx = p.ctx
		myCmdUint.dir = p.jobEntry.detail.WorkDir
		myCmdUint.user = p.jobEntry.detail.User
		myCmdUint.logName = fmt.Sprintf("%d.log", p.jobEntry.detail.ID)
		myCmdUint.logPath = p.logPath
		err = myCmdUint.launch()

		if err != nil {
			if isTimeout == false {
				p.jobEntry.detail.LastExitStatus = exitError
			}

			if p.jobEntry.detail.ErrorMailNotify {
				if err = rpcCall("Srv.SendMail", proto.SendMail{
					MailTo:  p.jobEntry.detail.MailTo,
					Subject: cfg.LocalAddr + "提醒脚本异常退出",
					Content: fmt.Sprintf(
						"任务名：%s\n详情：%v\n开始时间：%s\n异常：%s",
						p.jobEntry.detail.Name, p.jobEntry.detail.Commands, p.endTime.Format(proto.DefaultTimeLayout), err.Error()),
				}, &reply); err != nil {
					log.Error("Srv.SendMail error:", err, "server addr:", cfg.AdminAddr)
				}
			}

			if p.jobEntry.detail.ErrorAPINotify {
				postData, err := json.Marshal(proto.ApiPost{})
				if err != nil {
					log.Error("json.Marshal error:", err)
				}
				for _, url := range p.jobEntry.detail.APITo {
					if err = rpcCall("Srv.ApiPost", proto.ApiPost{
						Url:  url,
						Data: string(postData),
					}, &reply); err != nil {
						log.Error("Srv.ApiPost error:", err, "server addr:", cfg.AdminAddr)
					}
				}
			}
		}
		done = true

	}

	p.jobEntry.detail.LastCostTime = p.endTime.Sub(p.startTime).Seconds()
	if logContent != nil {
		*logContent = p.jobEntry.logContent
	}

	log.Infof("%s:%v %d %.3fs %v", p.jobEntry.detail.Name, p.jobEntry.detail.Commands, p.jobEntry.detail.Timeout, p.jobEntry.detail.LastCostTime, err)
}

type JobEntry struct {
	job    *crontab.Job
	detail models.CrontabJob
	// id         int
	ctx        context.Context
	cancel     context.CancelFunc
	processNum int32
	processes  map[int]*process
	pc         int32
	wg         util.WaitGroupWrapper
	ready      chan struct{}
	depends    []*depEntry
	logContent []byte
	jd         *Jiacrontabd
	mux        sync.RWMutex
	sync       bool
}

func newJobEntry(job *crontab.Job, jd *Jiacrontabd) *JobEntry {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	return &JobEntry{
		job:       job,
		cancel:    cancel,
		processes: make(map[int]*process),
		ctx:       ctx,
		jd:        jd,
	}
}

func (j *JobEntry) setPc() {
	atomic.AddInt32(&j.pc, 1)
}
func (j *JobEntry) getPc() int {
	return int(atomic.LoadInt32(&j.pc))
}

func (j *JobEntry) exec() []byte {

	j.wg.Wrap(func() {

		err := models.DB().Debug().Take(&j.detail, "id=? and status=?", j.job.ID, models.StatusJobTiming).Error
		if err != nil {
			log.Error("JobEntry.exec:", err)
			return
		}

		atomic.AddInt32(&j.processNum, 1)

		j.setPc()
		id := j.getPc()
		defer func() {
			atomic.AddInt32(&j.processNum, -1)
			models.DB().Model(&j.detail).Debug().Updates(map[string]interface{}{
				"status":           models.StatusJobTiming,
				"process_num":      gorm.Expr("process_num-?", 1),
				"last_cost_time":   j.detail.LastCostTime,
				"last_exit_status": j.detail.LastExitStatus,
			})
		}()

		models.DB().Model(&j.detail).Debug().Updates(map[string]interface{}{
			"status":         models.StatusJobRunning,
			"process_num":    gorm.Expr("process_num+?", 1),
			"last_exec_time": j.job.GetLastExecTime(),
			"next_exec_time": j.job.GetNextExecTime(),
		})

		p := newProcess(id, j)
		j.mux.Lock()
		j.processes[id] = p
		j.mux.Unlock()
		// 执行脚本
		p.exec(nil)
	})
	return nil
}

func (j *JobEntry) kill() {
	j.cancel()
	j.done()
	if err := models.DB().Model(&models.CrontabJob{}).Updates(map[string]interface{}{
		"status":      models.StatusJobStop,
		"process_num": 0,
	}).Error; err != nil {
		log.Error("JobEntry.kill", err)
	}
}

func (j *JobEntry) done() {
	select {
	case <-j.ctx.Done():
		j.mux.Lock()
		for _, v := range j.processes {
			v.cancel()
		}
		j.mux.Unlock()
		j.wg.Wait()
		log.Infof("job exit, ID:%d", j.job.ID)
	}
}

func (j *JobEntry) exit() {
	j.cancel()
}
