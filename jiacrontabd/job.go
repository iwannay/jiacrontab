package jiacrontabd

import (
	"context"
	"encoding/json"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/crontab"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/util"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/iwannay/log"
)

const (
	exitError       = "Error"
	exitKilled      = "Killed"
	exitSuccess     = "Success"
	exitDependError = "Dependent job execution failed"
	exitTimeout     = "Timeout"
)

type process struct {
	id        int
	deps      []*depEntry
	ctx       context.Context
	cancel    context.CancelFunc
	err       error
	startTime time.Time
	endTime   time.Time
	ready     chan struct{}
	retryNum  int
	jobEntry  *JobEntry
}

func newProcess(id int, jobEntry *JobEntry) *process {
	p := &process{
		id:        id,
		jobEntry:  jobEntry,
		startTime: time.Now(),
		ready:     make(chan struct{}),
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())

	for _, v := range p.jobEntry.detail.DependJobs {
		p.deps = append(p.deps, &depEntry{
			jobID:       p.jobEntry.detail.ID,
			processID:   id,
			jobUniqueID: p.jobEntry.uniqueID,
			id:          v.ID,
			from:        v.From,
			commands:    v.Commands,
			dest:        v.Dest,
			logPath:     filepath.Join(p.jobEntry.jd.getOpts().LogPath, "depend_job", time.Now().Format("2006/01/02"), fmt.Sprintf("%d-%s.log", v.JobID, v.ID)),
			done:        false,
			timeout:     v.Timeout,
		})
	}

	return p
}

func (p *process) waitDepExecDone() bool {

	if len(p.deps) == 0 {
		return true
	}

	ok := true
	if p.jobEntry.detail.IsSync {
		// 同步
		ok = p.jobEntry.jd.dispatchDependSync(p.ctx, p.deps, "")
	} else {
		// 并发模式
		ok = p.jobEntry.jd.dispatchDependAsync(p.ctx, p.deps)
	}
	if !ok {
		prefix := fmt.Sprintf("[%s %s] ", time.Now().Format("2006-01-02 15:04:05"), p.jobEntry.jd.getOpts().LocalAddr)
		p.jobEntry.logContent = append(p.jobEntry.logContent, []byte(prefix+"failed to exec depends, push depends error\n")...)
		return ok
	}

	c := time.NewTimer(3600 * time.Second)
	defer c.Stop()

	for {
		select {
		case <-p.ctx.Done():
			log.Debugf("jobID:%d exec cancel", p.jobEntry.detail.ID)
			return false
		case <-c.C:
			p.cancel()
			log.Errorf("jobID:%d exec dep timeout!", p.jobEntry.detail.ID)
			return false
		case <-p.ready:
			log.Debugf("jobID:%d exec all dep done.", p.jobEntry.detail.ID)
			return true
		}
	}
}

func (p *process) exec() error {
	var (
		ok       bool
		err      error
		doneChan = make(chan struct{}, 1)
	)

	if ok = p.waitDepExecDone(); !ok {
		p.jobEntry.handleDepError(p.startTime)
	} else {
		if p.jobEntry.detail.Timeout != 0 {
			time.AfterFunc(
				time.Duration(p.jobEntry.detail.Timeout)*time.Second, func() {
					select {
					case <-doneChan:
						close(doneChan)
					default:
						log.Debug("timeout callback:", "jobID:", p.jobEntry.detail.ID)
						p.jobEntry.timeoutTrigger(p)
					}
				})
		}

		myCmdUnit := cmdUint{
			args:             p.jobEntry.detail.Commands,
			ctx:              p.ctx,
			dir:              p.jobEntry.detail.WorkDir,
			user:             p.jobEntry.detail.WorkUser,
			env:              p.jobEntry.detail.WorkEnv,
			content:          p.jobEntry.logContent,
			logPath:          p.jobEntry.logPath,
			label:            p.jobEntry.detail.Name,
			killChildProcess: p.jobEntry.detail.KillChildProcess,
			jd:               p.jobEntry.jd,
		}

		if p.jobEntry.once {
			myCmdUnit.exportLog = true
		}
		p.err = myCmdUnit.launch()
		p.jobEntry.logContent = myCmdUnit.content

		doneChan <- struct{}{}

		if p.err != nil {
			p.jobEntry.handleNotify(p)
		}
	}

	p.endTime = time.Now()
	p.jobEntry.detail.LastCostTime = p.endTime.Sub(p.startTime).Seconds()

	log.Infof("%s exec %v %d %.3fs %v", p.jobEntry.detail.Name, p.jobEntry.detail.Commands, p.jobEntry.detail.Timeout, p.jobEntry.detail.LastCostTime, err)
	return p.err
}

type JobEntry struct {
	job        *crontab.Job
	detail     models.CrontabJob
	processNum int32
	processes  map[int]*process
	pc         int32
	wg         util.WaitGroupWrapper
	logContent []byte
	logPath    string
	jd         *Jiacrontabd
	mux        sync.RWMutex
	once       bool // 只执行一次
	sync       bool
	stop       int32 // job stop status
	uniqueID   string
}

func newJobEntry(job *crontab.Job, jd *Jiacrontabd) *JobEntry {
	return &JobEntry{
		uniqueID:  util.UUID(),
		job:       job,
		processes: make(map[int]*process),
		logPath:   filepath.Join(jd.getOpts().LogPath, "crontab_task", time.Now().Format("2006/01/02"), fmt.Sprintf("%d.log", job.ID)),
		jd:        jd,
	}
}

func (j *JobEntry) setPc() int {
	return int(atomic.AddInt32(&j.pc, 1))
}
func (j *JobEntry) getPc() int {
	return int(atomic.LoadInt32(&j.pc))
}

func (j *JobEntry) writeLog() {
	writeFile(j.logPath, &j.logContent)
}

func (j *JobEntry) handleDepError(startTime time.Time) {
	cfg := j.jd.getOpts()
	err := fmt.Errorf("%s %s execution of dependency job failed, jobID:%d", time.Now().Format(proto.DefaultTimeLayout), cfg.LocalAddr, j.detail.ID)
	endTime := time.Now()
	reply := true

	j.logContent = append(j.logContent, []byte(err.Error()+"\n")...)
	j.detail.LastExitStatus = exitDependError
	j.writeLog()

	if j.detail.ErrorMailNotify && len(j.detail.MailTo) != 0 {

		if err := j.jd.rpcCallCtx(context.TODO(), "Srv.SendMail", proto.SendMail{
			MailTo:  j.detail.MailTo,
			Subject: cfg.LocalAddr + "提醒脚本依赖异常退出",
			Content: fmt.Sprintf(
				"任务名：%s\n详情：%v\n开始时间：%s\n耗时：%.4f\n异常：%s",
				j.detail.Name, j.detail.Commands, endTime.Format(proto.DefaultTimeLayout), endTime.Sub(startTime).Seconds(), err),
		}, &reply); err != nil {
			log.Error("Srv.SendMail error:", err, "server addr:", cfg.AdminAddr)
		}
	}
}

func (j *JobEntry) handleNotify(p *process) {

	var (
		err   error
		reply bool
		cfg   = j.jd.getOpts()
	)

	if p.err == nil {
		return
	}

	if j.detail.ErrorMailNotify {
		if err = j.jd.rpcCallCtx(context.TODO(), "Srv.SendMail", proto.SendMail{
			MailTo:  j.detail.MailTo,
			Subject: cfg.LocalAddr + "提醒脚本异常退出",
			Content: fmt.Sprintf(
				"任务名：%s\n详情：%v\n开始时间：%s\n异常：%s\n重试次数：%d",
				j.detail.Name, j.detail.Commands,
				p.endTime.Format(proto.DefaultTimeLayout), p.err.Error(), p.retryNum),
		}, &reply); err != nil {
			log.Error("Srv.SendMail error:", err, "server addr:", cfg.AdminAddr)
		}
	}

	if j.detail.ErrorAPINotify {
		postData, err := json.Marshal(proto.CrontabApiNotifyBody{
			NodeAddr:  cfg.LocalAddr,
			JobName:   j.detail.Name,
			JobID:     int(j.detail.ID),
			Commands:  j.detail.Commands,
			CreatedAt: j.detail.CreatedAt,
			Timeout:   int64(j.detail.Timeout),
			Type:      "error",
			RetryNum:  p.retryNum,
		})
		if err != nil {
			log.Error("json.Marshal error:", err)
			return
		}

		if err = j.jd.rpcCallCtx(context.TODO(), "Srv.ApiPost", proto.ApiPost{
			Urls: j.detail.APITo,
			Data: string(postData),
		}, &reply); err != nil {
			log.Error("Srv.ApiPost error:", err, "server addr:", cfg.AdminAddr)
		}

	}
}

func (j *JobEntry) timeoutTrigger(p *process) {

	var (
		err   error
		reply bool
		cfg   = j.jd.getOpts()
	)

	for _, e := range j.detail.TimeoutTrigger {
		switch e {
		case proto.TimeoutTrigger_CallApi:
			j.detail.LastExitStatus = exitTimeout
			postData, err := json.Marshal(proto.CrontabApiNotifyBody{
				NodeAddr:  cfg.LocalAddr,
				JobName:   j.detail.Name,
				JobID:     int(j.detail.ID),
				Commands:  j.detail.Commands,
				CreatedAt: j.detail.CreatedAt,
				Timeout:   int64(j.detail.Timeout),
				Type:      "timeout",
				RetryNum:  p.retryNum,
			})
			if err != nil {
				log.Error("json.Marshal error:", err)
			}

			if err = j.jd.rpcCallCtx(context.TODO(), "Srv.ErrorNotify err:", proto.ApiPost{
				Urls: j.detail.APITo,
				Data: string(postData),
			}, &reply); err != nil {
				log.Error("Srv.ErrorNotify err:", err, "server addr:", cfg.AdminAddr)
			}

		case proto.TimeoutTrigger_SendEmail:
			j.detail.LastExitStatus = exitTimeout
			if err = j.jd.rpcCallCtx(context.TODO(), "Srv.SendMail", proto.SendMail{
				MailTo:  j.detail.MailTo,
				Subject: cfg.LocalAddr + "提醒脚本执行超时",
				Content: fmt.Sprintf(
					"任务名：%s\n详情：%v\n开始时间：%s\n超时：%ds\n重试次数：%d",
					j.detail.Name, j.detail.Commands, p.startTime.Format(proto.DefaultTimeLayout),
					j.detail.Timeout, p.retryNum),
			}, &reply); err != nil {
				log.Error("Srv.SendMail error:", err, "server addr:", cfg.AdminAddr)
			}
		case proto.TimeoutTrigger_Kill:
			j.detail.LastExitStatus = exitTimeout
			p.cancel()
		default:
			log.Error("invalid timeoutTrigger", e)
		}
	}
}

func (j *JobEntry) exec() {

	if atomic.LoadInt32(&j.stop) == 1 {
		return
	}

	j.wg.Wrap(func() {
		var err error
		if j.once {
			err = models.DB().Debug().Take(&j.detail, "id=?", j.job.ID).Error
		} else {
			err = models.DB().Debug().Take(&j.detail, "id=? and status in(?)",
				j.job.ID, []models.JobStatus{models.StatusJobTiming, models.StatusJobRunning}).Error
		}

		if err != nil {
			log.Info("JobEntry.exec:", err)
			return
		}

		if !j.once {
			if !j.detail.NextExecTime.Truncate(time.Second).Equal(j.job.GetNextExecTime().Truncate(time.Second)) {
				log.Error("JobEntry.exec time error:", j.detail.NextExecTime, "not equal", j.job.GetNextExecTime())
				return
			}
			j.jd.addJob(j.job)
		}
		if atomic.LoadInt32(&j.processNum) > int32(j.detail.MaxConcurrent) {
			return
		}

		atomic.AddInt32(&j.processNum, 1)

		id := j.setPc()
		startTime := time.Now()
		var endTime time.Time
		defer func() {
			endTime = time.Now()
			atomic.AddInt32(&j.processNum, -1)
			j.updateJob(models.StatusJobTiming, startTime, endTime, err)
		}()

		j.updateJob(models.StatusJobRunning, startTime, endTime, err)

		for i := 0; i <= j.detail.RetryNum; i++ {

			if atomic.LoadInt32(&j.stop) == 1 {
				return
			}

			log.Debug("jobID:", j.detail.ID, "retryNum:", i)

			p := newProcess(id, j)
			p.retryNum = i

			j.mux.Lock()
			j.processes[id] = p
			j.mux.Unlock()

			defer func() {
				j.mux.Lock()
				delete(j.processes, id)
				j.mux.Unlock()
			}()

			// 执行脚本
			if err = p.exec(); err == nil || j.once {
				break
			}
		}
	})
}

func (j *JobEntry) updateJob(status models.JobStatus, startTime, endTime time.Time, err error) {
	data := map[string]interface{}{
		"status":           status,
		"process_num":      atomic.LoadInt32(&j.processNum),
		"last_exec_time":   j.job.GetLastExecTime(),
		"last_exit_status": "",
	}

	if j.once && (status == models.StatusJobRunning) {
		data["process_num"] = gorm.Expr("process_num + ?", 1)
	}

	if j.once && (status == models.StatusJobTiming) {
		data["process_num"] = gorm.Expr("process_num - ?", 1)
	}

	var errMsg string
	if err != nil {
		errMsg = err.Error()
		data["last_exit_status"] = errMsg
	}

	if j.once {
		delete(data, "last_exec_time")
		delete(data, "status")
		delete(data, "last_exit_status")
	}

	if status == models.StatusJobTiming {
		if err = j.jd.rpcCallCtx(context.TODO(), "Srv.PushJobLog", models.JobHistory{
			JobType:   models.JobTypeCrontab,
			JobID:     j.detail.ID,
			Addr:      j.jd.getOpts().LocalAddr,
			JobName:   j.detail.Name,
			StartTime: startTime,
			EndTime:   endTime,
			ExitMsg:   errMsg,
		}, nil); err != nil {
			log.Error("rpc call Srv.PushJobLog failed:", err)
		}
	}

	models.DB().Model(&j.detail).Debug().Updates(data)
}

func (j *JobEntry) kill() {
	j.exit()
	if err := models.DB().Model(&j.detail).Updates(map[string]interface{}{
		"process_num": 0,
	}).Error; err != nil {
		log.Error("JobEntry.kill", err)
	}
}

func (j *JobEntry) waitDone() []byte {
	j.wg.Wait()
	atomic.StoreInt32(&j.stop, 0)
	return j.logContent
}

func (j *JobEntry) exit() {
	atomic.StoreInt32(&j.stop, 1)
	j.mux.Lock()
	for _, v := range j.processes {
		v.cancel()
	}
	j.mux.Unlock()
	j.waitDone()
}
