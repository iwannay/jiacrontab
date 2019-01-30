package crontab

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"
)

func Test_mine(t *testing.T) {
	fmt.Println(time.Now().Second())
	seconds := []int{5, 4, 100, 2, 34, 4}
	sort.Sort(sort.IntSlice(seconds))
	fmt.Println(seconds)
}

func Test_crontab_Ready(t *testing.T) {
	var timeLayout = "2006-01-02 15:04:05"
	c := New()
	now := time.Now().Add(6 * time.Second)
	c.AddTask(&Task{
		Value:    "海贼王" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})

	now = time.Now().Add(1 * time.Second)

	c.AddTask(&Task{
		Value:    "火影忍者" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})
	now = time.Now().Add(3 * time.Second)

	c.AddTask(&Task{
		Value:    "清风徐来" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})

	now = time.Now().Add(4 * time.Second)
	c.AddTask(&Task{
		Value:    "月上窗前" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})

	now = time.Now().Add(3 * time.Second)
	c.AddTask(&Task{
		Value:    "夕阳西下" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})

	bts, _ := json.MarshalIndent(c.GetAllTask(), "", "")
	fmt.Println(string(bts))

	go c.QueueScanWorker()

	go func() {
		for v := range c.Ready() {
			bts, _ := json.MarshalIndent(v, "", "")
			fmt.Println(string(bts))
		}
	}()

	time.Sleep(10 * time.Second)
}
