package main

import (
	"context"
	"fmt"
	"jiacrontab/client/store"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func newCrontab(taskChanSize int) *crontab {
	return &crontab{
		taskChan:     make(chan *proto.TaskArgs, taskChanSize),
		delTaskChan:  make(chan *proto.TaskArgs, taskChanSize),
		killTaskChan: make(chan *proto.TaskArgs, taskChanSize),
		handleMap:    make(map[string]*handle),
	}
}

type handle struct {
	cancel         context.CancelFunc   // 取消定时器
	cancelCmdArray []context.CancelFunc // 取消正在执行的脚本
	// resolvedDepends chan []byte
	readyDepends chan proto.MScriptContent
	clockChan    chan time.Time
	timeout      int64
}

type crontab struct {
	taskChan     chan *proto.TaskArgs
	delTaskChan  chan *proto.TaskArgs
	killTaskChan chan *proto.TaskArgs
	handleMap    map[string]*handle
	lock         sync.RWMutex
	sliceLock    sync.RWMutex
}

func (c *crontab) add(t *proto.TaskArgs) {
	c.taskChan <- t
}

func (c *crontab) quickStart(t *proto.TaskArgs, content *[]byte) {

	// if t.State != 0 {
	// 	log.Printf("quick start error task.State should 0, %d give", t.State)
	// } else {
	c.execTask(t, content)
	// }

	// var timeout int64
	// var err error
	// startTime := time.Now()
	// start := startTime.Unix()
	// args := strings.Split(t.Args, " ")
	// t.LastExecTime = start
	// if t.Timeout == 0 {
	// 	timeout = 600
	// } else {
	// 	timeout = t.Timeout
	// }
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	// // 垃圾回收
	// if t.State == 0 || t.State == 1 {
	// 	for k := range t.Depends {
	// 		t.Depends[k].Queue = make([]proto.MScriptContent, 0)
	// 	}
	// }

	// if ok := c.waitDependsDone(ctx, t.Id, &t.Depends, content, start, t.Sync); !ok {
	// 	cancel()
	// 	errMsg := fmt.Sprintf("[%s %s %s]>>  failded to exec depends\n", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, t.Name)
	// 	*content = append(*content, []byte(errMsg)...)
	// } else {
	// 	err = wrapExecScript(ctx, fmt.Sprintf("%s-%s.log", t.Name, t.Id), t.Command, globalConfig.logPath, content, args...)
	// 	cancel()
	// 	if err != nil {
	// 		*content = append(*content, []byte(err.Error())...)
	// 	}
	// }

	// t.LastCostTime = time.Now().UnixNano() - startTime.UnixNano()
	// globalStore.Sync()

	// log.Printf("%s:  quic start end costTime %ds %v", t.Name, t.LastCostTime, err)

}

// stop 停止计划任务
func (c *crontab) stop(t *proto.TaskArgs) {
	c.kill(t)
	c.delTaskChan <- t
}

func (c *crontab) kill(t *proto.TaskArgs) {
	c.killTaskChan <- t
}

func (c *crontab) delete(t *proto.TaskArgs) {
	globalStore.Update(func(s *store.Store) {
		delete(s.TaskList, t.Id)
	})
	c.kill(t)
	c.delTaskChan <- t
	log.Println("delete", t.Name, t.Id)
}

func (c *crontab) ids() []string {
	var sli []string
	c.lock.Lock()
	for k := range c.handleMap {
		sli = append(sli, k)
	}

	c.lock.Unlock()
	return sli
}

func (c *crontab) run() {
	// initialize
	go func() {
		globalStore.Update(func(s *store.Store) {
			for _, v := range s.TaskList {
				if v.State != 0 {

					c.add(v)
				}
			}
		}).Sync()

	}()
	// global clock
	go func() {
		t := time.Tick(1 * time.Minute)
		for {
			now := <-t

			// broadcast
			c.lock.Lock()
			for _, v := range c.handleMap {
				select {
				case v.clockChan <- now:
				case <-time.After(5 * time.Second):
				}

			}
			c.lock.Unlock()
		}
	}()

	// add task
	go func() {
		for {
			select {
			case task := <-c.taskChan:
				ctx, cancel := context.WithCancel(context.Background())
				c.lock.Lock()
				c.handleMap[task.Id] = &handle{
					cancel: cancel,
					// resolvedDepends: make(chan []byte),
					readyDepends: make(chan proto.MScriptContent, 10),
					clockChan:    make(chan time.Time),
				}
				c.lock.Unlock()
				task.State = 1
				log.Printf("add task %s %s", task.Name, task.Id)

				go c.deal(task, ctx)
			}
		}
	}()
	// remove task
	go func() {
		for {
			select {
			case task := <-c.delTaskChan:
				c.lock.Lock()
				if handle, ok := c.handleMap[task.Id]; ok {
					handle.cancel()
				}
				c.lock.Unlock()
			}
		}
	}()

	// kill task
	go func() {
		for {
			select {
			case task := <-c.killTaskChan:
				c.lock.Lock()
				if handle, ok := c.handleMap[task.Id]; ok {
					c.lock.Unlock()
					if handle.cancelCmdArray != nil {
						for _, cancel := range handle.cancelCmdArray {
							cancel()
						}
						handle.cancelCmdArray = make([]context.CancelFunc, 0)
						task.State = 1
						log.Println("kill", task.Name, task.Id)
						globalStore.Sync()
					}
				} else {
					c.lock.Unlock()
				}
			}
		}
	}()

}

func (c *crontab) deal(task *proto.TaskArgs, ctx context.Context) {
	var wgroup sync.WaitGroup
	for {
		c.lock.Lock()
		h := c.handleMap[task.Id]
		c.lock.Unlock()
		select {
		case now := <-h.clockChan:

			go func(now time.Time) {
				defer func() {
					libs.MRecover()
					wgroup.Done()
				}()

				check := task.C
				if checkMonth(check, now.Month()) &&
					checkWeekday(check, now.Weekday()) &&
					checkDay(check, now.Day()) &&
					checkHour(check, now.Hour()) &&
					checkMinute(check, now.Minute()) {
					var content []byte
					wgroup.Add(1)
					c.execTask(task, &content)
				}
			}(now)
		case <-ctx.Done():
			// 等待所有的计划任务执行完毕
			wgroup.Wait()
			task.State = 0
			log.Println("stop", task.Name, task.Id)
			// 垃圾回收
			// 防止向已终止的依赖接受通道继续发送信息
			for k := range task.Depends {
				task.Depends[k].Queue = make([]proto.MScriptContent, 0)
			}
			c.lock.Lock()
			close(c.handleMap[task.Id].clockChan)
			close(c.handleMap[task.Id].readyDepends)
			delete(c.handleMap, task.Id)
			c.lock.Unlock()
			globalStore.Sync()
			return
		}

	}

}

func (c *crontab) resolvedDepends(t *proto.TaskArgs, logContent []byte, taskTime int64, err string) {
	c.lock.Lock()
	if handle, ok := c.handleMap[t.Id]; ok {
		c.lock.Unlock()
		// handle.resolvedDepends <- logContent
		select {
		case handle.readyDepends <- proto.MScriptContent{
			TaskTime:   taskTime,
			LogContent: logContent,
			Err:        err,
			Done:       true,
		}:
		case <-time.After(5 * time.Second):
			log.Printf("taskTime %d failed to write to readyDepends chan", taskTime)
		}

	} else {
		c.lock.Unlock()
		log.Printf("depends: can not found %s", t.Id)
	}

}

func (c *crontab) waitDependsDone(ctx context.Context, taskId string, dpds *[]proto.MScript, logContent *[]byte, taskTime int64, sync bool) bool {
	defer func() {
		// 结束时修改执行状态
		if len(*dpds) > 0 {
			for k, v := range *dpds {
				for key, val := range v.Queue {
					if val.TaskTime == taskTime {
						(*dpds)[k].Queue[key].Done = true
					}
				}
			}
		}

	}()
	flag := true

	if len(*dpds) == 0 {
		log.Printf("taskId:%s depend length %d", taskId, len(*dpds))
		return true
	}

	// 一个脚本开始执行时把时间标志放入队列
	// 并显式声明执行未完成
	curQueueI := proto.MScriptContent{
		TaskTime: taskTime,
		Done:     false,
	}
	for k := range *dpds {
		(*dpds)[k].Queue = append((*dpds)[k].Queue, curQueueI)
	}

	syncFlag := true
	if sync {
		// 同步模式
		syncFlag = pushPipeDepend(*dpds, "", curQueueI)
	} else {
		// 并发模式
		// syncFlag = pushDepends(copyDpds)
		syncFlag = pushDepends(*dpds, curQueueI)
	}
	if !syncFlag {
		prefix := fmt.Sprintf("[%s %s]>>  ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr)
		*logContent = []byte(prefix + "failed to exec depends push depends error!\n")
		return syncFlag
	}

	c.lock.Lock()
	// 任务在停止状态下需要手动构造依赖接受通道
	if _, ok := c.handleMap[taskId]; !ok {
		c.handleMap[taskId] = &handle{
			readyDepends: make(chan proto.MScriptContent, 10),
		}
	}

	if handle, ok := c.handleMap[taskId]; ok {
		c.lock.Unlock()

		// 默认所有依赖最终总超时3600
		t := time.Tick(3600 * time.Second)
		for {

			select {
			case <-ctx.Done():
				return false
			case <-t:
				log.Printf("failed to exec depends wait timeout!")
				return false
			// case *logContent = <-handle.resolvedDepends:
			case val := <-handle.readyDepends:
				if val.TaskTime != taskTime {

					checkFlag := false
					for _, v := range (*dpds)[0].Queue {
						if v.TaskTime == val.TaskTime {
							checkFlag = true
						}
					}

					if checkFlag {
						handle.readyDepends <- val
						log.Printf("task %s depend<%d> return to readyDepends chan", taskId, val.TaskTime)
						// 防止重复接收
						time.Sleep(1 * time.Second)
					}

				} else {
					*logContent = val.LogContent
					if val.Err != "" {
						flag = false
					}
					goto end
				}

			}
		}

	} else {
		c.lock.Unlock()
		log.Printf("depends: can not found task %s", taskId)
		return false
	}
end:
	log.Printf("task:%s exec all depends done", taskId)
	return flag
}

func (c *crontab) execTask(task *proto.TaskArgs, logContent *[]byte) {
	var err error

	now2 := time.Now()
	start := now2.UnixNano()
	args := strings.Split(task.Args, " ")
	task.LastExecTime = now2.Unix()
	saveState := task.State
	task.State = 2
	atomic.AddInt32(&task.NumberProcess, 1)
	ctx, cancel := context.WithCancel(context.Background())

	// 保存并发执行的终止句柄
	c.lock.Lock()
	if hdl, ok := c.handleMap[task.Id]; ok {
		c.lock.Unlock()
		if len(hdl.cancelCmdArray) >= task.MaxConcurrent {
			hdl.cancelCmdArray[0]()
			hdl.cancelCmdArray = hdl.cancelCmdArray[1:]
		}
		hdl.cancelCmdArray = append(hdl.cancelCmdArray, cancel)
	} else {
		c.lock.Unlock()
	}

	log.Printf("start task %s %s %s %s", task.Name, task.Id, task.Command, task.Args)

	if ok := c.waitDependsDone(ctx, task.Id, &task.Depends, logContent, now2.Unix(), task.Sync); !ok {
		cancel()
		errMsg := fmt.Sprintf("[%s %s %s]>>  failded to exec depends\n", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, task.Name)
		*logContent = append(*logContent, []byte(errMsg)...)
		writeLog(globalConfig.logPath, fmt.Sprintf("%s-%s.log", task.Name, task.Id), logContent)
		if task.UnexpectedExitMail {
			costTime := time.Now().UnixNano() - start
			sendMail(task.MailTo, globalConfig.addr+"提醒脚本依赖异常退出", fmt.Sprintf(
				"任务名：%s\n详情：%s %v\n开始时间：%s\n耗时：%.4f\n异常：%s",
				task.Name, task.Command, task.Args, now2.Format("2006-01-02 15:04:05"), float64(costTime)/1000000000, err.Error()))
		}

	} else {
		flag := true

		if task.Timeout != 0 {
			time.AfterFunc(time.Duration(task.Timeout)*time.Second, func() {
				if flag {
					switch task.OpTimeout {
					case "email":
						sendMail(task.MailTo, globalConfig.addr+"提醒脚本执行超时", fmt.Sprintf(
							"任务名：%s\n详情：%s %v\n开始时间：%s\n超时：%ds",
							task.Name, task.Command, task.Args, now2.Format("2006-01-02 15:04:05"), task.Timeout))
					case "kill":
						cancel()

					case "email_and_kill":
						cancel()
						sendMail(task.MailTo, globalConfig.addr+"提醒脚本执行超时", fmt.Sprintf(
							"任务名：%s\n详情：%s %v\n开始时间：%s\n超时：%ds",
							task.Name, task.Command, task.Args, now2.Format("2006-01-02 15:04:05"), task.Timeout))
					case "ignore":
					default:
					}
				}

			})
		}

		err = wrapExecScript(ctx, fmt.Sprintf("%s-%s.log", task.Name, task.Id), task.Command, globalConfig.logPath, logContent, args...)

		flag = false
		if err != nil && task.UnexpectedExitMail {
			sendMail(task.MailTo, globalConfig.addr+"提醒脚本异常退出", fmt.Sprintf(
				"任务名：%s\n详情：%s %v\n开始时间：%s\n异常：%s",
				task.Name, task.Command, task.Args, now2.Format("2006-01-02 15:04:05"), err.Error()))

		}
	}
	atomic.AddInt32(&task.NumberProcess, -1)
	task.LastCostTime = time.Now().UnixNano() - start
	if task.NumberProcess == 0 {
		if saveState != 2 {
			task.State = saveState
		} else {
			task.State = 1
		}

		// 垃圾回收
		c.sliceLock.Lock()
		for k := range task.Depends {
			task.Depends[k].Queue = make([]proto.MScriptContent, 0)
		}
		c.sliceLock.Unlock()

	} else {
		task.State = 2
	}
	globalStore.Sync()

	log.Printf("%s:%s %v %s %.3fs %v", task.Name, task.Command, task.Args, task.OpTimeout, float64(task.LastCostTime)/1000000000, err)
}
