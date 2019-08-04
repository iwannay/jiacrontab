// Package pqueue jiacrontab中使用的优先队列
// 参考nsq的实现
// 做了注释和少量调整
package pqueue

import (
	"container/heap"
)

type Item struct {
	Value    interface{}
	Priority int64
	Index    int
}

// PriorityQueue 最小堆实现的优先队列
type PriorityQueue []*Item

// New 创建
func New(capacity int) PriorityQueue {
	return make(PriorityQueue, 0, capacity)
}

// Less 队列长队
func (pq PriorityQueue) Len() int {
	return len(pq)
}

// Less 比较相邻两个原素优先级
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

// Swap 交换相邻原素
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

// Push 添加新的item
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	c := cap(*pq)
	if n+1 > c {
		npq := make(PriorityQueue, n, c*2)
		copy(npq, *pq)
		*pq = npq
	}
	*pq = (*pq)[0 : n+1]
	item := x.(*Item)
	item.Index = n
	(*pq)[n] = item
}

func (pq *PriorityQueue) update(item *Item, value string, priority int64) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.Index)
}

// Pop 弹出队列末端原素
func (pq *PriorityQueue) Pop() interface{} {
	n := len(*pq)
	c := cap(*pq)
	if n < (c/2) && c > 25 {
		npq := make(PriorityQueue, n, c/2)
		copy(npq, *pq)
		*pq = npq
	}
	item := (*pq)[n-1]
	item.Index = -1
	*pq = (*pq)[0 : n-1]
	return item
}

// PeekAndShift 根据比较max并弹出原素
func (pq *PriorityQueue) PeekAndShift(max int64) (*Item, int64) {
	if pq.Len() == 0 {
		return nil, 0
	}

	item := (*pq)[0]
	if item.Priority > max {
		return nil, item.Priority - max
	}
	heap.Remove(pq, 0)
	return item, 0
}
