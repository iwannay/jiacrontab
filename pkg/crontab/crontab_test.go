package crontab

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func Test_crontab_Ready(t *testing.T) {
	var timeLayout = "2006-01-02 15:04:05"
	c := New()
	now := time.Now().Add(6 * time.Second)
	c.AddTask(&Task{
		Value:    "test1" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})

	now = time.Now().Add(1 * time.Second)

	c.AddTask(&Task{
		Value:    "test2" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})
	now = time.Now().Add(3 * time.Second)

	c.AddTask(&Task{
		Value:    "test3" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})

	now = time.Now().Add(4 * time.Second)
	c.AddTask(&Task{
		Value:    "test4" + now.Format(timeLayout),
		Priority: now.UnixNano(),
	})

	now = time.Now().Add(3 * time.Second)
	c.AddTask(&Task{
		Value:    "test5" + now.Format(timeLayout),
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
