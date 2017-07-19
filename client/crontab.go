package main

import (
	"context"
	"errors"
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
}

func (c *crontab) add(t *proto.TaskArgs) {
	c.taskChan <- t
	log.Printf("add task %s %s", t.Name, t.Id)
}

func (c *crontab) quickStart(t *proto.TaskArgs, content *[]byte) {
	var timeout int64
	var err error
	startTime := time.Now()
	start := startTime.Unix()
	args := strings.Split(t.Args, " ")
	t.LastExecTime = start
	if t.Timeout == 0 {
		timeout = 600
	} else {
		timeout = t.Timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	// 垃圾回收
	if t.State == 0 || t.State == 1 {
		for k := range t.Depends {
			t.Depends[k].Queue = make([]proto.MScriptContent, 0)
		}
	}

	if ok := c.waitDependsDone(ctx, t.Id, &t.Depends, content, start); !ok {
		cancel()
		err = errors.New("failded to exec depends")
		*content = append(*content, []byte(err.Error())...)
	} else {
		err = execScript(ctx, fmt.Sprintf("%s-%s.log", t.Name, t.Id), t.Command, globalConfig.logPath, content, args...)
		cancel()
		if err != nil {
			*content = append(*content, []byte(err.Error())...)
		}
	}

	t.LastCostTime = time.Now().UnixNano() - startTime.UnixNano()
	globalStore.Sync()

	log.Printf("%s:  quic start end costTime %ds %v", t.Name, t.LastCostTime, err)

}

// stop 停止计划任务
func (c *crontab) stop(t *proto.TaskArgs) {
	c.kill(t)
	c.delTaskChan <- t
	log.Println("stop", t.Name, t.Id)
}

func (c *crontab) kill(t *proto.TaskArgs) {
	c.killTaskChan <- t
	log.Println("kill", t.Name, t.Id)
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
				v.clockChan <- now
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
				task.State = 1
				c.lock.Lock()
				c.handleMap[task.Id] = &handle{
					cancel: cancel,
					// resolvedDepends: make(chan []byte),
					readyDepends: make(chan proto.MScriptContent, 10),
					clockChan:    make(chan time.Time),
				}
				c.lock.Unlock()

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
			wgroup.Add(1)
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
					var hdl *handle
					var err error

					// 垃圾回收
					if task.State == 0 || task.State == 1 {
						for k := range task.Depends {
							task.Depends[k].Queue = make([]proto.MScriptContent, 0)
						}
					}

					now2 := time.Now()
					start := now2.UnixNano()
					args := strings.Split(task.Args, " ")
					task.LastExecTime = now2.Unix()
					task.State = 2
					atomic.AddInt32(&task.NumberProcess, 1)
					ctx, cancel := context.WithCancel(context.Background())

					// 保存并发执行的终止句柄
					c.lock.Lock()
					hdl = c.handleMap[task.Id]
					if len(hdl.cancelCmdArray) >= task.MaxConcurrent {
						hdl.cancelCmdArray[0]()
						hdl.cancelCmdArray = hdl.cancelCmdArray[1:]
					}
					hdl.cancelCmdArray = append(hdl.cancelCmdArray, cancel)
					c.lock.Unlock()

					log.Printf("start task %s %s %s %s", task.Name, task.Id, task.Command, task.Args)

					if ok := c.waitDependsDone(ctx, task.Id, &task.Depends, &content, now2.Unix()); !ok {
						cancel()
						err = errors.New("failded to exec depends")
						content = append(content, []byte(err.Error())...)
						costTime := time.Now().UnixNano() - start
						sendMail(task.MailTo, globalConfig.addr+"提醒脚本依赖超时退出", fmt.Sprintf(
							"任务名：%s\n详情：%s %v\n开始时间：%s\n耗时：%.4f\n异常:%s",
							task.Name, task.Command, task.Args, now2.Format("2006-01-02 15:04:05"), float64(costTime)/1000000000, err.Error()))
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

						err = execScript(ctx, fmt.Sprintf("%s-%s.log", task.Name, task.Id), task.Command, globalConfig.logPath, &content, args...)

						flag = false
						if err != nil {
							sendMail(task.MailTo, globalConfig.addr+"提醒脚本异常退出", fmt.Sprintf(
								"任务名：%s\n详情：%s %v\n开始时间：%s\n异常：%s",
								task.Name, task.Command, task.Args, now2.Format("2006-01-02 15:04:05"), err.Error()))
						}
					}
					atomic.AddInt32(&task.NumberProcess, -1)
					task.LastCostTime = time.Now().UnixNano() - start
					if task.NumberProcess == 0 {
						task.State = 1
					} else {
						task.State = 2
					}
					globalStore.Sync()

					log.Printf("%s:%s %v %s %.3fs %v", task.Name, task.Command, task.Args, task.OpTimeout, float64(task.LastCostTime)/1000000000, err)

				}
			}(now)
		case <-ctx.Done():
			// 等待所有的计划任务执行完毕
			wgroup.Wait()
			task.State = 0
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

func (c *crontab) resolvedDepends(t *proto.TaskArgs, logContent []byte, taskTime int64) {
	c.lock.Lock()
	if handle, ok := c.handleMap[t.Id]; ok {
		c.lock.Unlock()
		// handle.resolvedDepends <- logContent
		handle.readyDepends <- proto.MScriptContent{
			TaskTime:   taskTime,
			LogContent: logContent,
			Done:       true,
		}
	} else {
		c.lock.Unlock()
		log.Printf("depends: can not found %s", t.Id)
	}

}

func (c *crontab) waitDependsDone(ctx context.Context, taskId string, dpds *[]proto.MScript, logContent *[]byte, taskTime int64) bool {
	// defer func() {
	// 	for k, _ := range *dpds {
	// 		(*dpds)[k].Done = false
	// 		(*dpds)[k].LogContent = []byte("")
	// 	}
	// }()

	if len(*dpds) == 0 {
		log.Printf("taskId:%s dpend length %d", taskId, len(*dpds))
		return true
	}

	// 一个脚本开始执行时把时间标志放入队列
	// 并显式声明执行未完成
	for k, _ := range *dpds {
		(*dpds)[k].Queue = append((*dpds)[k].Queue, proto.MScriptContent{
			TaskTime: taskTime,
			Done:     false,
		})
	}
	if ok := pushDepends(taskId, *dpds); !ok {
		*logContent = []byte("failed to exec depends push depends error!\n")
		return false
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

		// 默认所有依赖最终欧超时1小时
		t := time.Tick(600 * time.Second)
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

					handle.readyDepends <- val
					log.Printf("task %s depend<%d> return to readyDepends chan", taskId, val.TaskTime)
					// 防止重复接受
					time.Sleep(1 * time.Second)
				} else {
					*logContent = val.LogContent
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
	return true
}
