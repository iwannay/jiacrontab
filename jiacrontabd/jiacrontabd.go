package jiacrontabd

import (
	"jiacrontab/models"
	"jiacrontab/pkg/crontab"
	"jiacrontab/pkg/finder"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"

	"jiacrontab/pkg/util"
	"sync"
	"time"
)

// Jiacrontabd scheduling center
type Jiacrontabd struct {
	crontab *crontab.Crontab
	// All jobs added
	jobs            map[uint]*JobEntry
	dep             *dependencies
	daemon          *daemon
	heartbeatPeriod time.Duration
	mux             sync.RWMutex
	wg              util.WaitGroupWrapper
}

// New return a Jiacrontabd instance
func New() *Jiacrontabd {
	j := &Jiacrontabd{
		jobs:            make(map[uint]*JobEntry),
		daemon:          newDaemon(100),
		heartbeatPeriod: 5 * time.Second,
		crontab:         crontab.New(),
	}
	j.dep = newDependencies(j)

	return j
}

func (j *Jiacrontabd) addJob(job *crontab.Job) {
	j.mux.Lock()
	if _, ok := j.jobs[job.ID]; ok {
		j.mux.Unlock()
	} else {
		j.jobs[job.ID] = newJobEntry(job, j)
		j.mux.Unlock()
	}
	if t, err := job.NextExecutionTime(time.Now()); err != nil {
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
	job.NextExecutionTime(time.Now())
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
	j.wg.Wrap(j.crontab.QueueScanWorker)
	for v := range j.crontab.Ready() {
		v := v.Value.(*crontab.Job)
		log.Info("job queue:", v)
		j.execTask(v)
	}
}

// filterDepend 本地执行的脚本依赖不再请求网络，直接转发到对应的处理模块
// 目标网络不是本机时返回false
func (j *Jiacrontabd) filterDepend(task *depEntry) bool {
	if task.dest != cfg.LocalAddr {
		return false
	}

	isAllDone := true
	j.mux.Lock()
	if h, ok := j.jobs[task.jobID]; ok {
		j.mux.Unlock()

		var logContent []byte
		var curTaskEntry *process

		for _, v := range h.processes {
			if v.id == task.processID {
				curTaskEntry = v
				for _, vv := range v.jobEntry.depends {
					if vv.done == false {
						isAllDone = false
					} else {
						logContent = append(logContent, vv.logContent...)
					}

					if vv.saveID == task.saveID && v.jobEntry.sync {
						if ok := j.pushPipeDepend(v.jobEntry.depends, vv.saveID); ok {
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

		// 如果依赖脚本执行出错直接通知主脚本停止
		if task.err != nil {
			isAllDone = true
			log.Infof("depend %s %s exec failed %s try to stop master task", task.name, task.commands, task.err)
		}

		if isAllDone {
			curTaskEntry.jobEntry.ready <- struct{}{}
			curTaskEntry.jobEntry.logContent = logContent
		}

	} else {
		log.Infof("cannot find task handler %s %s", task.name, task.commands)
		j.mux.Unlock()
	}

	return true

}

func (j *Jiacrontabd) pushPipeDepend(deps []*depEntry, depEntryID string) bool {
	var flag = true
	if len(deps) > 0 {
		flag = false
		l := len(deps) - 1
		for k, v := range deps {
			if flag || depEntryID == "" {
				// 检测目标服务器为本机时直接执行脚本
				log.Infof("sync push %s %s", v.dest, v.commands)
				if v.dest == cfg.LocalAddr {
					j.dep.add(v)
				} else {
					var reply bool
					err := rpcCall("Srv.Depends", []proto.DepJob{{
						ID:        v.saveID,
						Name:      v.name,
						Dest:      v.dest,
						From:      v.from,
						JobID:     v.jobID,
						ProcessID: v.processID,
						Commands:  v.commands,
						Timeout:   v.timeout,
					}}, &reply)
					if !reply || err != nil {
						log.Error("Logic.Depends error:", err, "server addr:", cfg.AdminAddr)
						return false
					}
				}
				flag = true
				break
			}

			if (v.saveID == depEntryID) && (l != k) {
				flag = true
			}

		}

	}
	return flag
}

func (j *Jiacrontabd) pushDepend(deps []*depEntry) bool {
	var depJobs proto.DepJobs
	for _, v := range deps {
		// 检测目标服务器是本机直接执行脚本
		if v.dest == cfg.LocalAddr {
			j.dep.add(v)
		} else {
			depJobs = append(depJobs, proto.DepJob{
				ID:        v.saveID,
				Name:      v.name,
				Dest:      v.dest,
				From:      v.from,
				ProcessID: v.processID,
				JobID:     v.jobID,
				Commands:  v.commands,
				Timeout:   v.timeout,
			})
		}
	}

	if len(depJobs) > 0 {
		var reply bool
		if err := rpcCall("Srv.Depend", depJobs, &reply); err != nil {
			log.Error("Srv.Depends error:", err, "server addr:", cfg.AdminAddr)
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
	err := rpcCall(rpc.RegisterService, models.Node{
		Addr:           cfg.LocalAddr,
		DaemonTaskNum:  j.daemon.count(),
		CrontabTaskNum: j.count(),
		Name:           hostname,
	}, &reply)

	if err != nil {
		log.Error("Srv.Register error:", err, ",server addr:", cfg.AdminAddr)
	}

	time.AfterFunc(heartbeatPeriod, j.heartBeat)
}

func (j *Jiacrontabd) init() {
	models.CreateDB(cfg.DriverName, cfg.DSN)
	models.DB().CreateTable(&models.CrontabJob{}, &models.DaemonJob{})
	models.DB().AutoMigrate(&models.CrontabJob{}, &models.DaemonJob{})
	if cfg.AutoCleanTaskLog {
		go finder.SearchAndDeleteFileOnDisk(cfg.LogPath, 24*time.Hour*30, 1<<30)
	}
}

// Main main function
func (j *Jiacrontabd) Main() {
	j.init()
	j.heartBeat()
	go j.run()
	rpc.ListenAndServe(cfg.ListenAddr, newCrontabJobSrv(j), newDaemonJobSrv(j), &Srv{})
}
