package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"jiacrontab/client/store"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	exitError       = "error"
	exitKilled      = "killed"
	exitSuccess     = "success"
	exitDependError = "depend error"
	exitTimeout     = "timeout"
)

type taskEntity struct {
	id         string
	pid        string
	name       string
	command    string
	args       string
	taskArgs   *proto.TaskArgs
	state      int
	timeout    int64
	sync       bool
	cancel     context.CancelFunc
	logContent []byte
	ready      chan struct{}
	depends    []*dependScript
}

type dependScript struct {
	pid        string
	id         string
	from       string
	command    string
	args       string
	dest       string
	done       bool
	timeout    int64
	err        error
	name       string
	logContent []byte
}

func newTaskEntity(t *proto.TaskArgs) *taskEntity {
	var depends []*dependScript
	var dependSubName string
	var md5Sum string
	id := fmt.Sprintf("%d", time.Now().Unix())

	for k, v := range t.Depends {
		md5Sum = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d", v.Command, v.Args, k))))
		if v.Name == "" {
			dependSubName = md5Sum
		} else {
			dependSubName = v.Name
		}
		depends = append(depends, &dependScript{
			pid:     id,
			id:      fmt.Sprintf("%s-%s-%s", t.Id, id, md5Sum),
			from:    v.From,
			dest:    v.Dest,
			command: v.Command,
			timeout: v.Timeout,
			args:    v.Args,
			name:    fmt.Sprintf("%s-%s", t.Name, dependSubName),
			done:    false,
		})
	}
	return &taskEntity{
		id:       id,
		pid:      t.Id,
		name:     t.Name,
		command:  t.Command,
		sync:     t.Sync,
		taskArgs: t,
		ready:    make(chan struct{}),
		depends:  depends,
	}
}

func (t *taskEntity) exec(logContent *[]byte) {
	var err error
	now := time.Now()
	atomic.AddInt32(&t.taskArgs.NumberProcess, 1)
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	start := now.UnixNano()
	t.taskArgs.LastExecTime = now.Unix()
	t.taskArgs.State = 2
	t.taskArgs.LastExitStatus = exitSuccess
	flag := true
	isExceptError := false

	if ok := t.waitDependsDone(ctx); !ok {
		cancel()
		errMsg := fmt.Sprintf("[%s %s %s]>>  Execution of dependency script failed\n", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, t.name)
		t.logContent = append(t.logContent, []byte(errMsg)...)
		writeLog(globalConfig.logPath, fmt.Sprintf("%s.log", t.name), &t.logContent)
		t.taskArgs.LastExitStatus = exitDependError
		isExceptError = true
		if t.taskArgs.UnexpectedExitMail {
			costTime := time.Now().UnixNano() - start
			sendMail(t.taskArgs.MailTo, globalConfig.addr+"提醒脚本依赖异常退出", fmt.Sprintf(
				"任务名：%s\n详情：%s %v\n开始时间：%s\n耗时：%.4f\n异常：%s",
				t.name, t.taskArgs.Command, t.taskArgs.Args, now.Format("2006-01-02 15:04:05"), float64(costTime)/1000000000, errors.New(errMsg)))
		}

	} else {
		// 执行脚本
		if t.taskArgs.Timeout != 0 {
			time.AfterFunc(time.Duration(t.taskArgs.Timeout)*time.Second, func() {
				if flag {

					isExceptError = true
					switch t.taskArgs.OpTimeout {
					case "email":
						t.taskArgs.LastExitStatus = exitTimeout
						sendMail(t.taskArgs.MailTo, globalConfig.addr+"提醒脚本执行超时", fmt.Sprintf(
							"任务名：%s\n详情：%s %v\n开始时间：%s\n超时：%ds",
							t.taskArgs.Name, t.taskArgs.Command, t.taskArgs.Args, now.Format("2006-01-02 15:04:05"), t.taskArgs.Timeout))
					case "kill":
						t.taskArgs.LastExitStatus = exitTimeout
						cancel()

					case "email_and_kill":
						t.taskArgs.LastExitStatus = exitTimeout
						cancel()
						sendMail(t.taskArgs.MailTo, globalConfig.addr+"提醒脚本执行超时", fmt.Sprintf(
							"任务名：%s\n详情：%s %v\n开始时间：%s\n超时：%ds",
							t.taskArgs.Name, t.taskArgs.Command, t.taskArgs.Args, now.Format("2006-01-02 15:04:05"), t.taskArgs.Timeout))
					case "ignore":
					default:
					}
				}

			})
		}

		var cmdList [][]string
		cmd := []string{t.taskArgs.Command, t.taskArgs.Args}
		cmdList = append(cmdList, cmd)
		if len(t.taskArgs.PipeCommands) > 0 {
			cmdList = append(cmdList, t.taskArgs.PipeCommands...)
		}

		err := wrapExecScript(ctx, fmt.Sprintf("%s.log", t.name), cmdList, globalConfig.logPath, &t.logContent)
		flag = false

		if err != nil {
			if isExceptError == false {
				t.taskArgs.LastExitStatus = exitError
			}

			if t.taskArgs.UnexpectedExitMail {
				sendMail(t.taskArgs.MailTo, globalConfig.addr+"提醒脚本异常退出", fmt.Sprintf(
					"任务名：%s\n详情：%s %v\n开始时间：%s\n异常：%s",
					t.taskArgs.Name, t.taskArgs.Command, t.taskArgs.Args, now.Format("2006-01-02 15:04:05"), err.Error()))
			}
		}
	}

	atomic.AddInt32(&t.taskArgs.NumberProcess, -1)

	t.taskArgs.LastCostTime = time.Now().UnixNano() - start

	if t.taskArgs.TimerCounter > 0 {
		if t.taskArgs.NumberProcess == 0 {
			t.taskArgs.State = 1
		}
	} else {
		t.taskArgs.State = 0
	}
	globalStore.Sync()

	if logContent != nil {
		*logContent = t.logContent
	}

	log.Printf("%s:%s %v %s %.3fs %v", t.taskArgs.Name, t.taskArgs.Command, t.taskArgs.Args, t.taskArgs.OpTimeout, float64(t.taskArgs.LastCostTime)/1000000000, err)

}

type handle struct {
	cancel    context.CancelFunc // 取消定时器
	clockChan chan time.Time
	taskPool  []*taskEntity
}

type crontab struct {
	taskChan     chan *proto.TaskArgs
	stopTaskChan chan *proto.TaskArgs
	killTaskChan chan *proto.TaskArgs
	handleMap    map[string]*handle
	lock         sync.RWMutex
}

func newCrontab(taskChanSize int) *crontab {
	return &crontab{
		taskChan:     make(chan *proto.TaskArgs, taskChanSize),
		stopTaskChan: make(chan *proto.TaskArgs, taskChanSize),
		killTaskChan: make(chan *proto.TaskArgs, taskChanSize),
		handleMap:    make(map[string]*handle),
	}
}

func (c *crontab) add(t *proto.TaskArgs) {
	c.taskChan <- t
}

func (c *crontab) quickStart(t *proto.TaskArgs, content *[]byte) {
	taskEty := newTaskEntity(t)
	c.lock.Lock()
	if _, ok := c.handleMap[t.Id]; !ok {

		taskPool := make([]*taskEntity, 0)
		c.handleMap[t.Id] = &handle{
			taskPool: append(taskPool, taskEty),
		}
		c.lock.Unlock()
	} else {
		c.handleMap[t.Id].taskPool = append(c.handleMap[t.Id].taskPool, taskEty)
		c.lock.Unlock()
	}

	taskEty.exec(content)
}

// stop 停止计划任务并杀死正在执行的脚本进程
func (c *crontab) stop(t *proto.TaskArgs) {
	c.kill(t)
	c.stopTaskChan <- t
}

// 杀死正在执行的脚本进程
func (c *crontab) kill(t *proto.TaskArgs) {
	c.killTaskChan <- t
}

// 删除计划任务
func (c *crontab) delete(t *proto.TaskArgs) {
	globalStore.Update(func(s *store.Store) {
		delete(s.TaskList, t.Id)
	})
	c.stop(t)
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
			for k, v := range c.handleMap {
				if v.clockChan == nil {
					log.Printf("clock:%s is closed", k)
					continue
				}
				select {
				case v.clockChan <- now:
				case <-time.After(1 * time.Second):
				}

			}
			c.lock.Unlock()
		}
	}()

	// add task
	go func() {
		for {
			select {
			case t := <-c.taskChan:
				c.lock.Lock()
				if h, ok := c.handleMap[t.Id]; !ok {
					ctx, cancel := context.WithCancel(context.Background())
					taskPool := make([]*taskEntity, 0)
					c.handleMap[t.Id] = &handle{
						cancel:    cancel,
						clockChan: make(chan time.Time),
						taskPool:  taskPool,
					}
					c.lock.Unlock()
					go c.deal(t, ctx)
					log.Printf("add task %s %s", t.Name, t.Id)
				} else {
					if h.cancel == nil {
						ctx, cancel := context.WithCancel(context.Background())
						c.handleMap[t.Id].cancel = cancel
						if c.handleMap[t.Id].clockChan == nil {
							c.handleMap[t.Id].clockChan = make(chan time.Time)
						}
						c.lock.Unlock()
						go c.deal(t, ctx)
					} else {
						c.lock.Unlock()
					}
					log.Printf("task %s %s exists", t.Name, t.Id)
				}

			}
		}
	}()
	// stop task crontab
	go func() {
		for {
			select {
			case task := <-c.stopTaskChan:
				c.lock.Lock()
				if handle, ok := c.handleMap[task.Id]; ok {
					if handle.cancel != nil {
						handle.cancel()
						log.Printf("try to stop timer %s", task.Name)
					} else {
						log.Printf("stop timer %s  failed cancel function is nil", task.Name)
					}

				} else {
					log.Printf("stop: can not found timer %s", task.Name)
					task.State = 0
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
					if handle.taskPool != nil {
						for k, v := range handle.taskPool {
							if v.cancel == nil {
								log.Println("kill", task.Name, task.Id, k, "but cancel handler is nul")
							} else {
								v.cancel()
								log.Println("kill", task.Name, task.Id, k)
							}
						}

					}
				} else {
					log.Printf("kill: can not found %s", task.Name)
					c.lock.Unlock()
				}
			}
		}
	}()

}

func (c *crontab) deal(task *proto.TaskArgs, ctx context.Context) {
	var wgroup sync.WaitGroup
	// 定时计数器用于统计有多少个定时期，当定时器为0时说明没有正在执行的计划
	atomic.AddInt32(&task.TimerCounter, 1)
	task.State = 1
	defer atomic.AddInt32(&task.TimerCounter, -1)
	c.lock.Lock()
	h := c.handleMap[task.Id]
	c.lock.Unlock()
	for {

		select {
		case now := <-h.clockChan:

			go func(now time.Time) {
				defer func() {
					libs.MRecover()
					wgroup.Done()
				}()

				wgroup.Add(1)
				check := task.C
				if checkMonth(check, now.Month()) &&
					checkWeekday(check, now.Weekday()) &&
					checkDay(check, now.Day()) &&
					checkHour(check, now.Hour()) &&
					checkMinute(check, now.Minute()) {

					taskEty := newTaskEntity(task)
					h.taskPool = append(h.taskPool, taskEty)
					if l := len(h.taskPool); l > task.MaxConcurrent {
						cancelPool := h.taskPool[0 : l-task.MaxConcurrent]
						h.taskPool = h.taskPool[l-task.MaxConcurrent:]
						for k, v := range cancelPool {
							if v.cancel != nil {
								v.cancel()
								log.Printf("taskPool: clean %s %d", v.name, k)
							}
						}
					}
					taskEty.exec(nil)
				}
			}(now)
		case <-ctx.Done():
			// 等待所有的计划任务执行完毕
			wgroup.Wait()
			c.lock.Lock()
			task.State = 0
			if task.NumberProcess == 0 {
				close(c.handleMap[task.Id].clockChan)
				delete(c.handleMap, task.Id)
			} else {
				close(c.handleMap[task.Id].clockChan)
				c.handleMap[task.Id].cancel = nil
			}

			c.lock.Unlock()
			log.Printf("stop %s %s ok", task.Name, task.Id)
			globalStore.Sync()
			return
		}

	}

}

func (t *taskEntity) waitDependsDone(ctx context.Context) bool {
	if len(t.depends) == 0 {
		log.Printf("%s depend length %d", t.name, 0)
		return true
	}

	syncFlag := true
	if t.sync {
		// 同步模式
		syncFlag = pushPipeDepend(t.depends, "")
	} else {
		syncFlag = pushDepends(t.depends)
	}
	if !syncFlag {
		prefix := fmt.Sprintf("[%s %s]>>  ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr)
		t.logContent = append(t.logContent, []byte(prefix+"failed to exec depends push depends error!\n")...)
		return syncFlag
	}

	// 默认所有依赖最终总超时3600
	c := time.NewTimer(3600 * time.Second)
	for {
		select {
		case <-ctx.Done():
			c.Stop()
			return false
		case <-c.C:
			log.Printf("%s failed to exec depends wait timeout!", t.name)
			c.Stop()
			return false
		case <-t.ready:
			c.Stop()
			log.Printf("%s exec all depends done", t.name)
			return true
		}
	}
}
