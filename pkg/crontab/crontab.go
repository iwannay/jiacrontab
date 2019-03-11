package crontab

import (
	"container/heap"
	"jiacrontab/pkg/pqueue"
	"sync"
	"time"
)

// Task 任务
type Task = pqueue.Item

type Crontab struct {
	pq    pqueue.PriorityQueue
	mux   sync.RWMutex
	ready chan *Task
}

func New() *Crontab {
	return &Crontab{
		pq:    pqueue.New(100),
		ready: make(chan *Task, 10000),
	}
}

// AddJob 添加未经处理的job
func (c *Crontab) AddJob(j *Job) error {
	nt, err := j.NextExecutionTime(time.Now())
	if err != nil {
		return err
	}

	heap.Push(&c.pq, &Task{
		Priority: nt.UnixNano(),
		Value:    j,
	})
	return nil
}

// AddJob 添加延时任务
func (c *Crontab) AddTask(t *Task) {
	heap.Push(&c.pq, t)
}

func (c *Crontab) Len() int {
	c.mux.RLock()
	len := len(c.pq)
	c.mux.RUnlock()
	return len
}

func (c *Crontab) GetAllTask() []*Task {
	c.mux.Lock()
	list := c.pq
	c.mux.Unlock()
	return list
}

func (c *Crontab) Ready() <-chan *Task {
	return c.ready
}

func (c *Crontab) QueueScanWorker() {
	refreshTicker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-refreshTicker.C:
			if len(c.pq) == 0 {
				continue
			}
			c.mux.Lock()
			now := time.Now().UnixNano()
			job, _ := c.pq.PeekAndShift(now)
			c.mux.Unlock()
			if job == nil {
				continue
			}
			c.ready <- job
		}
	}
}
