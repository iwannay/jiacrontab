package jiacrontabd

import (
	"context"
	"jiacrontab/pkg/crontab"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/util"
	"math/rand"
	"sync"
	"sync/atomic"
)

type process struct {
	id     int
	cancel context.CancelFunc
	*JobEntry
}

func newProcess(id int, job *JobEntry) *process {
	return &process{
		id:       id,
		JobEntry: job,
	}
}

type JobEntry struct {
	job        *crontab.Job
	id         int
	ctx        context.Context
	cancel     context.CancelFunc
	processNum int32
	cancels    []context.CancelFunc
	processes  map[int]*process
	wg         util.WaitGroupWrapper
	ready      chan struct{}
	depends    []*depEntry
	logContent []byte
	mux        sync.RWMutex
	sync       bool
}

func newJobEntry(job *crontab.Job) *JobEntry {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	return &JobEntry{
		job:       job,
		cancel:    cancel,
		processes: make(map[int]*process),
		ctx:       ctx,
	}
}

func (j *JobEntry) exec() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	j.cancels = append(j.cancels, cancel)
	atomic.AddInt32(&j.processNum, 1)
	j.mux.Lock()
	id := rand.Int()
	j.processes[id] = newProcess(id, j)
	j.mux.Unlock()
	j.wg.Wrap(func() {
		defer atomic.AddInt32(&j.processNum, -1)
		// 执行脚本
	})
}

func (j *JobEntry) done() {
	select {
	case <-j.ctx.Done():
		for _, v := range j.cancels {
			v()
		}
		j.wg.Wait()
		log.Infof("job exit, ID:%d", j.job.ID)
	}
}

func (j *JobEntry) exit() {
	j.cancel()
}
