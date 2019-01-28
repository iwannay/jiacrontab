package crontab

import (
	"container/heap"
	"jiacrontab/pkg/pqueue"
	"sync"
	"time"
)

// Task 任务
type Task = pqueue.Item

type crontab struct {
	pq    pqueue.PriorityQueue
	mux   sync.RWMutex
	ready chan *Task
}

func New() *crontab {
	return &crontab{
		pq:    pqueue.New(100),
		ready: make(chan *Task, 10000),
	}
}

func (c *crontab) Add(t *Task) {
	heap.Push(&c.pq, t)
}

func (c *crontab) Len() int {
	c.mux.RLock()
	len := len(c.pq)
	c.mux.RUnlock()
	return len
}

func (c *crontab) GetAllTask() []*Task {
	c.mux.Lock()
	list := c.pq
	c.mux.Unlock()
	return list
}

func (c *crontab) Ready() <-chan *Task {
	return c.ready
}

func (c *crontab) QueueScanWorker() {
	refreshTicker := time.NewTicker(100 * time.Millisecond)
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
		default:
		}
	}
}
