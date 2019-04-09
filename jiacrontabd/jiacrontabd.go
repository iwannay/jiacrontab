package jiacrontabd

import (
	"jiacrontab/models"
	"jiacrontab/pkg/crontab"
	"jiacrontab/pkg/finder"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"

	"github.com/iwannay/log"

	"jiacrontab/pkg/util"
	"sync"
	"time"
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
	wg              util.WaitGroupWrapper
}

// New return a Jiacrontabd instance
func New() *Jiacrontabd {
	j := &Jiacrontabd{
		jobs:            make(map[uint]*JobEntry),
		tmpJobs:         make(map[string]*JobEntry),
		daemon:          newDaemon(100),
		heartbeatPeriod: 5 * time.Second,
		crontab:         crontab.New(),
	}
	j.dep = newDependencies(j)

	return j
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

func (j *Jiacrontabd) addJob(job *crontab.Job) {
	j.mux.Lock()
	if v, ok := j.jobs[job.ID]; ok {
		v.job = job
	} else {
		j.jobs[job.ID] = newJobEntry(job, j)
	}
	j.mux.Unlock()

	if t, err := job.NextExecutionTime(job.GetNextExecTime()); err != nil {
		log.Error("NextExecutionTime:", err, " timeArgs:", job)
	} else {
		if err := models.DB().Model(&models.CrontabJob{}).Where("id=?", job.ID).Debug().
			Updates(map[string]interface{}{
				"next_exec_time": t,
				"status":         models.StatusJobTiming,
			}).Error; err != nil {
			log.Error(err)
		}

		j.crontab.AddJob(job)
	}
}

func (j *Jiacrontabd) execTask(job *crontab.Job) {

	j.mux.RLock()
	if task, ok := j.jobs[job.ID]; ok {
		j.mux.RUnlock()
		task.exec()
		return
	}
	j.mux.RUnlock()

}

func (j *Jiacrontabd) killTask(jobID uint) {
	j.mux.RLock()
	if task, ok := j.jobs[jobID]; ok {
		j.mux.RUnlock()
		task.kill()
		return
	}
	j.mux.RUnlock()
}

func (j *Jiacrontabd) run() {
	j.dep.run()
	j.daemon.run()
	j.wg.Wrap(j.crontab.QueueScanWorker)

	for v := range j.crontab.Ready() {
		v := v.Value.(*crontab.Job)
		log.Info("job queue:", v)
		j.execTask(v)
	}
}

// SetDependDone 依赖执行完毕时设置相关状态
// 目标网络不是本机时返回false
func (j *Jiacrontabd) SetDependDone(task *depEntry) bool {

	if task.dest != cfg.LocalAddr {
		return false
	}

	isAllDone := true
	j.mux.Lock()
	job, ok := j.jobs[task.jobID]
	if !ok {
		job, ok = j.tmpJobs[task.jobUniqueID]
	}
	j.mux.Unlock()

	if ok && task.jobUniqueID == job.uniqueID {

		var logContent []byte
		var curTaskEntry *process

		for _, p := range job.processes {
			if p.id == task.processID {
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

					if dep.id == task.id && p.jobEntry.sync {
						if ok := j.dispatchDependSync(p.deps, dep.id); ok {
							return true
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
			log.Infof("depend %s %s exec failed, %s, try to stop master task", task.name, task.commands, task.err)
		}

		if isAllDone {
			curTaskEntry.ready <- struct{}{}
			curTaskEntry.jobEntry.logContent = logContent
		}

	} else {
		log.Infof("cannot find task handler %s %s", task.name, task.commands)
		j.mux.Unlock()
	}

	return true

}

// 同步模式根据depEntryID确定位置实现任务的依次调度
func (j *Jiacrontabd) dispatchDependSync(deps []*depEntry, depEntryID string) bool {
	flag := true
	if len(deps) > 0 {
		flag = false
		l := len(deps) - 1
		for k, v := range deps {
			// 根据flag实现调度下一个依赖任务
			if flag || depEntryID == "" {
				// 检测目标服务器为本机时直接执行脚本
				if v.dest == cfg.LocalAddr {
					j.dep.add(v)
				} else {
					var reply bool
					err := rpcCall("Srv.ExecDepend", []proto.DepJob{{
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
						log.Error("Srv.ExecDepend error:", err, "server addr:", cfg.AdminAddr)
						return false
					}
				}
				flag = true
				break
			}

			if (v.id == depEntryID) && (l != k) {
				flag = true
			}

		}

	}
	return flag
}

func (j *Jiacrontabd) dispatchDependAsync(deps []*depEntry) bool {
	var depJobs proto.DepJobs
	for _, v := range deps {
		// 检测目标服务器是本机直接执行脚本
		if v.dest == cfg.LocalAddr {
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
		if err := rpcCall("Srv.ExecDepend", depJobs, &reply); err != nil {
			log.Error("Srv.ExecDepend error:", err, "server addr:", cfg.AdminAddr)
			return false
		}
	}

	return true
}

func (j *Jiacrontabd) count() int {
	j.mux.RLock()
	num := len(j.jobs)
	j.mux.RUnlock()
	return num
}

func (j *Jiacrontabd) heartBeat() {
	var reply bool
	hostname := cfg.Hostname
	if hostname == "" {
		hostname = util.GetHostname()
	}
	node := models.Node{
		Addr:           cfg.LocalAddr,
		DaemonTaskNum:  j.daemon.count(),
		CrontabTaskNum: j.count(),
		Name:           hostname,
	}

	models.DB().Model(&models.CrontabJob{}).Where("status=?", models.StatusJobUnaudited).Count(&node.CrontabJobAuditNum)
	models.DB().Model(&models.DaemonJob{}).Where("status=?", models.StatusJobUnaudited).Count(&node.DaemonJobAuditNum)

	err := rpcCall(rpc.RegisterService, node, &reply)

	if err != nil {
		log.Error("Srv.Register error:", err, ",server addr:", cfg.AdminAddr)
	}

	time.AfterFunc(heartbeatPeriod, j.heartBeat)
}

func (j *Jiacrontabd) recovery() {
	var crontabJobs []models.CrontabJob
	var daemonJobs []models.DaemonJob

	err := models.DB().Debug().Find(&crontabJobs, "status in (?)", []models.JobStatus{models.StatusJobTiming, models.StatusJobRunning}).Error
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
		})
	}

	err = models.DB().Find(&daemonJobs, "status in (?)", []models.JobStatus{models.StatusJobOk, models.StatusJobStop}).Error

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
	models.CreateDB(cfg.DriverName, cfg.DSN)
	models.DB().CreateTable(&models.CrontabJob{}, &models.DaemonJob{})
	models.DB().AutoMigrate(&models.CrontabJob{}, &models.DaemonJob{})
	j.startTime = time.Now()
	if cfg.AutoCleanTaskLog {
		go finder.SearchAndDeleteFileOnDisk(cfg.LogPath, 24*time.Hour*30, 1<<30)
	}
}

// Main main function
func (j *Jiacrontabd) Main() {
	j.init()
	j.heartBeat()
	go j.run()
	rpc.ListenAndServe(cfg.ListenAddr, newCrontabJobSrv(j), newDaemonJobSrv(j), newSrv(j))
}
