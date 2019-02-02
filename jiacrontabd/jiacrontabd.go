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
	jobs            map[int]*JobEntry
	dep             *dependencies
	daemon          *daemon
	heartbeatPeriod time.Duration
	mux             sync.RWMutex
	wg              util.WaitGroupWrapper
}

// New return a Jiacrontabd instance
func New() *Jiacrontabd {
	j := &Jiacrontabd{
		jobs:            make(map[int]*JobEntry),
		daemon:          newDaemon(100),
		heartbeatPeriod: 5 * time.Second,
	}
	j.dep = newDependencies(j)

	return j
}

// AddJob add job
func (j *Jiacrontabd) AddJob(job *crontab.Job) {
	j.mux.Lock()
	if _, ok := j.jobs[job.ID]; ok {
		j.mux.Unlock()
		return
	}

	j.jobs[job.ID] = newJobEntry(job)
	j.mux.Unlock()
	j.crontab.AddJob(job)
}

func (j *Jiacrontabd) execTask(job *crontab.Job) {
	j.mux.RLock()
	if task, ok := j.jobs[job.ID]; !ok {
		j.mux.RUnlock()
		task.exec()
		return
	}
}

// Run start Jiacrontabd instance
func (j *Jiacrontabd) run() {
	// tcp server
	j.wg.Wrap(j.crontab.QueueScanWorker)
	for v := range j.crontab.Ready() {
		v := v.Value.(*crontab.Job)
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
				for _, vv := range v.depends {
					if vv.done == false {
						isAllDone = false
					} else {
						logContent = append(logContent, vv.logContent...)
					}

					if vv.id == task.id && v.sync {
						if ok := j.pushPipeDepend(v.depends, vv.id); ok {
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
			curTaskEntry.ready <- struct{}{}
			curTaskEntry.logContent = logContent
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
					err := rpcCall("Logic.Depends", []proto.DependsTask{{
						ID:         v.id,
						Name:       v.name,
						Dest:       v.dest,
						From:       v.from,
						JobEntryID: v.jobID,
						ProcessID:  v.processID,
						Commands:   v.commands,
						Timeout:    v.timeout,
					}}, &reply)
					if !reply || err != nil {
						log.Error("Logic.Depends error:", err, "server addr:", cfg.AdminAddr)
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

func (j *Jiacrontabd) count() int {
	j.mux.RLock()
	num := len(j.jobs)
	j.mux.RUnlock()
	return num
}

func (j *Jiacrontabd) heartBeat() {
	var mail proto.MailArgs
	err := rpcCall(rpc.RegisterService, models.Client{
		Addr:           cfg.LocalAddr,
		DaemonTaskNum:  j.daemon.count(),
		CrontabTaskNum: j.count(),
		State:          1,
		Mail:           cfg.MailTo,
	}, &mail)

	if err != nil {
		log.Error("Logic.Register error:", err, "server addr:", cfg.AdminAddr)
	}

	time.AfterFunc(heartbeatPeriod, j.heartBeat)
}

// Main main function
func (j *Jiacrontabd) Main() {

	if cfg.AutoCleanTaskLog {
		go finder.SearchAndDeleteFileOnDisk(cfg.LogPath, 24*time.Hour*30, 1<<30)
	}

	j.heartBeat()
	rpc.ListenAndServe(cfg.ListenAddr)
}
