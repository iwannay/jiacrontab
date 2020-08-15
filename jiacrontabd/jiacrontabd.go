package jiacrontabd

import (
	"context"
	"jiacrontab/models"
	"jiacrontab/pkg/crontab"
	"jiacrontab/pkg/finder"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"sync/atomic"

	"github.com/iwannay/log"

	"jiacrontab/pkg/util"
	"sync"
	"time"

	"fmt"
	"strings"
)

// Jiacrontabd scheduling center
type Jiacrontabd struct {
	crontab *crontab.Crontab
	// All jobs added
	jobs            map[uint]*JobEntry
	tmpJobs         map[string]*JobEntry
	dep             *dependencies
	daemon          *Daemon
	heartbeatPeriod time.Duration
	mux             sync.RWMutex
	startTime       time.Time
	cfg             atomic.Value
	wg              util.WaitGroupWrapper
}

// New return a Jiacrontabd instance
func New(opt *Config) *Jiacrontabd {
	j := &Jiacrontabd{
		jobs:    make(map[uint]*JobEntry),
		tmpJobs: make(map[string]*JobEntry),

		heartbeatPeriod: 5 * time.Second,
		crontab:         crontab.New(),
	}
	j.swapOpts(opt)
	j.dep = newDependencies(j)
	j.daemon = newDaemon(100, j)

	return j
}

func (j *Jiacrontabd) getOpts() *Config {
	return j.cfg.Load().(*Config)
}

func (j *Jiacrontabd) swapOpts(opts *Config) {
	j.cfg.Store(opts)
}

func (j *Jiacrontabd) addTmpJob(job *JobEntry) {
	j.mux.Lock()
	j.tmpJobs[job.uniqueID] = job
	j.mux.Unlock()
}

func (j *Jiacrontabd) removeTmpJob(job *JobEntry) {
	j.mux.Lock()
	delete(j.tmpJobs, job.uniqueID)
	j.mux.Unlock()
}

func (j *Jiacrontabd) addJob(job *crontab.Job, updateLastExecTime bool) error {
	j.mux.Lock()
	if v, ok := j.jobs[job.ID]; ok {
		v.job = job
	} else {
		var crontabJob models.CrontabJob
		err := models.DB().First(&crontabJob, "id=?", job.ID).Error
		if err != nil {
			log.Error(err)
			j.mux.Unlock()
			return nil
		}
		if len(crontabJob.WorkIp) > 0 && !checkIpInWhiteList(strings.Join(crontabJob.WorkIp, ",")) {
			if err := models.DB().Model(&models.CrontabJob{}).Where("id=?", job.ID).
				Updates(map[string]interface{}{
					"status":         models.StatusJobStop,
					"next_exec_time": time.Time{},
					"lastExitStatus": "IP受限制",
				}).Error; err != nil {
				log.Error(err)
			}
			j.mux.Unlock()
			return nil
		}
		j.jobs[job.ID] = newJobEntry(job, j)
	}
	j.mux.Unlock()

	if err := j.crontab.AddJob(job); err != nil {
		log.Error("NextExecutionTime:", err, " timeArgs:", job)
		return fmt.Errorf("时间格式错误: %v - %s", err, job.Format())
	}
	data := map[string]interface{}{
		"next_exec_time": job.GetNextExecTime(),
		"status":         models.StatusJobTiming,
	}

	if updateLastExecTime {
		data["last_exec_time"] = time.Now()
	}

	if err := models.DB().Model(&models.CrontabJob{}).Where("id=?", job.ID).
		Updates(data).Error; err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (j *Jiacrontabd) execTask(job *crontab.Job) {

	j.mux.RLock()
	if task, ok := j.jobs[job.ID]; ok {
		j.mux.RUnlock()
		task.exec()
		return
	}
	log.Errorf("not found jobID(%d)", job.ID)
	j.mux.RUnlock()

}

func (j *Jiacrontabd) killTask(jobID uint) {
	var jobs []*JobEntry
	j.mux.RLock()
	if job, ok := j.jobs[jobID]; ok {
		jobs = append(jobs, job)
	}

	for _, v := range j.tmpJobs {
		if v.detail.ID == jobID {
			jobs = append(jobs, v)
		}
	}
	j.mux.RUnlock()

	for _, v := range jobs {
		v.kill()
	}
}

func (j *Jiacrontabd) run() {
	j.dep.run()
	j.daemon.run()
	j.wg.Wrap(j.crontab.QueueScanWorker)

	for v := range j.crontab.Ready() {
		v := v.Value.(*crontab.Job)
		j.execTask(v)
	}
}

// SetDependDone 依赖执行完毕时设置相关状态
// 目标网络不是本机时返回false
func (j *Jiacrontabd) SetDependDone(task *depEntry) bool {

	var (
		ok  bool
		job *JobEntry
	)

	if task.dest != j.getOpts().BoardcastAddr {
		return false
	}

	isAllDone := true

	j.mux.Lock()
	if job, ok = j.tmpJobs[task.jobUniqueID]; !ok {
		job, ok = j.jobs[task.jobID]
	}
	j.mux.Unlock()
	if ok {

		var logContent []byte
		var curTaskEntry *process

		for _, p := range job.processes {
			if int(p.id) == task.processID {
				curTaskEntry = p
				for _, dep := range p.deps {

					if dep.id == task.id {
						dep.dest = task.dest
						dep.from = task.from
						dep.logContent = task.logContent
						dep.err = task.err
						dep.done = true
					}

					if dep.done == false {
						isAllDone = false
					} else {
						logContent = append(logContent, dep.logContent...)
					}
					// 同步模式上一个依赖结束才会触发下一个
					if dep.id == task.id && task.err == nil && p.jobEntry.detail.IsSync {
						if err := j.dispatchDependSync(p.ctx, p.deps, dep.id); err != nil {
							task.err = err
						}
					}

				}
			}
		}

		if curTaskEntry == nil {
			log.Infof("cannot find task entry %s %s", task.name, task.commands)
			return true
		}

		// 如果依赖任务执行出错直接通知主任务停止
		if task.err != nil {
			isAllDone = true
			curTaskEntry.err = task.err
			log.Infof("depend %s %s exec failed, %s, try to stop master task", task.name, task.commands, task.err)
		}

		if isAllDone {
			curTaskEntry.ready <- struct{}{}
			curTaskEntry.jobEntry.logContent = append(curTaskEntry.jobEntry.logContent, logContent...)
		}

	} else {
		log.Infof("cannot find task handler %s %s", task.name, task.commands)
	}

	return true

}

// 同步模式根据depEntryID确定位置实现任务的依次调度
func (j *Jiacrontabd) dispatchDependSync(ctx context.Context, deps []*depEntry, depEntryID string) error {
	flag := true
	cfg := j.getOpts()
	if len(deps) > 0 {
		flag = false
		for _, v := range deps {
			// 根据flag实现调度下一个依赖任务
			if flag || depEntryID == "" {
				// 检测目标服务器为本机时直接执行脚本
				if v.dest == cfg.BoardcastAddr {
					j.dep.add(v)
				} else {
					var reply bool
					err := j.rpcCallCtx(ctx, "Srv.ExecDepend", []proto.DepJob{{
						ID:          v.id,
						Name:        v.name,
						Dest:        v.dest,
						From:        v.from,
						JobUniqueID: v.jobUniqueID,
						JobID:       v.jobID,
						ProcessID:   v.processID,
						Commands:    v.commands,
						Timeout:     v.timeout,
					}}, &reply)
					if !reply || err != nil {
						return fmt.Errorf("Srv.ExecDepend error:%v server addr:%s", err, cfg.AdminAddr)
					}
				}
				break
			}

			if v.id == depEntryID {
				flag = true
			}

		}
	}
	return nil
}

func (j *Jiacrontabd) dispatchDependAsync(ctx context.Context, deps []*depEntry) error {
	var depJobs proto.DepJobs
	cfg := j.getOpts()
	for _, v := range deps {
		// 检测目标服务器是本机直接执行脚本
		if v.dest == cfg.BoardcastAddr {
			j.dep.add(v)
		} else {
			depJobs = append(depJobs, proto.DepJob{
				ID:          v.id,
				Name:        v.name,
				Dest:        v.dest,
				From:        v.from,
				ProcessID:   v.processID,
				JobID:       v.jobID,
				JobUniqueID: v.jobUniqueID,
				Commands:    v.commands,
				Timeout:     v.timeout,
			})
		}
	}

	if len(depJobs) > 0 {
		var reply bool
		if err := j.rpcCallCtx(ctx, "Srv.ExecDepend", depJobs, &reply); err != nil {
			return fmt.Errorf("Srv.ExecDepend error:%v server addr: %s", err, cfg.AdminAddr)

		}
	}
	return nil
}

func (j *Jiacrontabd) count() int {
	j.mux.RLock()
	num := len(j.jobs)
	j.mux.RUnlock()
	return num
}

func (j *Jiacrontabd) deleteJob(jobID uint) {
	j.mux.Lock()
	delete(j.jobs, jobID)
	j.mux.Unlock()
}

func (j *Jiacrontabd) heartBeat() {
	var (
		reply    bool
		cronJobs []struct {
			Total   uint
			GroupID uint
			Failed  bool
			Status  models.JobStatus
		}
		daemonJobs []struct {
			Total   uint
			GroupID uint
			Status  models.JobStatus
		}
		ok             bool
		nodes          = make(map[uint]models.Node)
		cfg            = j.getOpts()
		nodeName       = cfg.NodeName
		node           models.Node
		superGroupNode models.Node
	)

	if nodeName == "" {
		nodeName = util.GetHostname()
	}

	models.DB().Model(&models.CrontabJob{}).Select("id,group_id,status,failed,count(1) as total").Group("group_id,status,failed").Scan(&cronJobs)
	models.DB().Model(&models.DaemonJob{}).Select("id,group_id,status,count(1) as total").Group("group_id,status").Scan(&daemonJobs)

	nodes[models.SuperGroup.ID] = models.Node{
		Addr:    cfg.BoardcastAddr,
		GroupID: models.SuperGroup.ID,
		Name:    nodeName,
	}

	for _, job := range cronJobs {
		superGroupNode = nodes[models.SuperGroup.ID]
		if node, ok = nodes[job.GroupID]; !ok {
			node = models.Node{
				Addr:    cfg.BoardcastAddr,
				GroupID: job.GroupID,
				Name:    nodeName,
			}
		}

		if job.Failed && job.Status == models.StatusJobTiming || job.Status == models.StatusJobRunning {
			node.CrontabJobFailNum += job.Total
			superGroupNode.CrontabJobFailNum += job.Total
		}

		if job.Status == models.StatusJobUnaudited {
			node.CrontabJobAuditNum += job.Total
			superGroupNode.CrontabJobAuditNum += job.Total
		}

		if job.Status == models.StatusJobTiming || job.Status == models.StatusJobRunning {
			node.CrontabTaskNum += job.Total
			superGroupNode.CrontabTaskNum += job.Total
		}

		nodes[job.GroupID] = node
		nodes[models.SuperGroup.ID] = superGroupNode
	}

	for _, job := range daemonJobs {
		superGroupNode = nodes[models.SuperGroup.ID]
		if node, ok = nodes[job.GroupID]; !ok {
			node = models.Node{
				Addr:    cfg.BoardcastAddr,
				GroupID: job.GroupID,
				Name:    nodeName,
			}
		}
		if job.Status == models.StatusJobUnaudited {
			node.DaemonJobAuditNum += job.Total
			superGroupNode.DaemonJobAuditNum += job.Total
		}
		if job.Status == models.StatusJobRunning {
			node.DaemonTaskNum += job.Total
			superGroupNode.DaemonTaskNum += job.Total
		}
		nodes[job.GroupID] = node
		nodes[models.SuperGroup.ID] = superGroupNode
	}

	err := j.rpcCallCtx(context.TODO(), rpc.RegisterService, nodes, &reply)

	if err != nil {
		log.Error("Srv.Register error:", err, ",server addr:", cfg.AdminAddr)
	}

	time.AfterFunc(time.Duration(j.getOpts().ClientAliveInterval)*time.Second, j.heartBeat)
}

func (j *Jiacrontabd) recovery() {
	var crontabJobs []models.CrontabJob
	var daemonJobs []models.DaemonJob

	err := models.DB().Find(&crontabJobs, "status IN (?)", []models.JobStatus{models.StatusJobTiming, models.StatusJobRunning}).Error
	if err != nil {
		log.Debug("crontab recovery:", err)
	}

	for _, v := range crontabJobs {
		j.addJob(&crontab.Job{
			ID:      v.ID,
			Second:  v.TimeArgs.Second,
			Minute:  v.TimeArgs.Minute,
			Hour:    v.TimeArgs.Hour,
			Day:     v.TimeArgs.Day,
			Month:   v.TimeArgs.Month,
			Weekday: v.TimeArgs.Weekday,
		}, false)
	}

	err = models.DB().Find(&daemonJobs, "status in (?)", []models.JobStatus{models.StatusJobOk}).Error

	if err != nil {
		log.Debug("daemon recovery:", err)
	}

	for _, v := range daemonJobs {
		job := v
		j.daemon.add(&daemonJob{
			job: &job,
		})
	}

}

func (j *Jiacrontabd) init() {
	cfg := j.getOpts()
	if err := models.CreateDB(cfg.DriverName, cfg.DSN); err != nil {
		panic(err)
	}
	models.DB().AutoMigrate(&models.CrontabJob{}, &models.DaemonJob{})
	j.startTime = time.Now()
	if cfg.AutoCleanTaskLog {
		go finder.SearchAndDeleteFileOnDisk(cfg.LogPath, 24*time.Hour*30, 1<<30)
	}
	j.recovery()
}

func (j *Jiacrontabd) rpcCallCtx(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	return rpc.CallCtx(j.getOpts().AdminAddr, serviceMethod, ctx, args, reply)
}

// Main main function
func (j *Jiacrontabd) Main() {
	j.init()
	j.heartBeat()
	go j.run()
	rpc.ListenAndServe(j.getOpts().ListenAddr, newCrontabJobSrv(j), newDaemonJobSrv(j), newSrv(j))
}
