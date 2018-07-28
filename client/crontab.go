package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"jiacrontab/libs/proto"
	"jiacrontab/model"
	"log"
	"path/filepath"
	"strings"
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
	pid        uint
	name       string
	command    string
	args       string
	logPath    string
	taskArgs   *model.CrontabTask
	state      int
	timeout    int64
	sync       bool
	cancel     context.CancelFunc
	logContent []byte
	ready      chan struct{}
	depends    []*dependScript
}

type dependScript struct {
	taskId       uint   // 定时任务id
	taskEntityId string // 当前依赖的父级任务（可能存在多个并发的task）id
	id           string // 当前依赖id
	from         string
	command      string
	args         string
	dest         string
	done         bool
	timeout      int64
	err          error
	name         string
	logContent   []byte
}

func newTaskEntity(t *model.CrontabTask) *taskEntity {
	var depends []*dependScript
	var dependSubName string
	var md5Sum string
	id := fmt.Sprintf("%d_%d", t.ID, time.Now().Unix())

	for k, v := range t.Depends {
		md5Sum = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d", v.Command, v.Args, k))))
		if v.Name == "" {
			dependSubName = md5Sum
		} else {
			dependSubName = v.Name
		}
		depends = append(depends, &dependScript{
			taskEntityId: id,
			taskId:       t.ID,
			id:           fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d-%s", k, dependSubName)))),
			from:         v.From,
			dest:         v.Dest,
			command:      v.Command,
			timeout:      v.Timeout,
			args:         v.Args,
			name:         fmt.Sprintf("%d-%s", k, dependSubName),
			done:         false,
		})
	}
	return &taskEntity{
		id:       id,
		pid:      t.ID,
		name:     t.Name,
		command:  t.Command,
		sync:     t.Sync,
		taskArgs: t,
		logPath:  filepath.Join(globalConfig.logPath, "crontab_task"),
		ready:    make(chan struct{}),
		depends:  depends,
	}
}

func (t *taskEntity) exec(logContent *[]byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("%s exec panic %s \n", t.taskArgs.Name, err)
		}
		model.DB().Model(&model.CrontabTask{}).Where("id=?", t.taskArgs.ID).Update(map[string]interface{}{
			"state":            t.taskArgs.State,
			"last_cost_time":   t.taskArgs.LastCostTime,
			"last_exec_time":   t.taskArgs.LastExecTime,
			"last_exit_status": t.taskArgs.LastExitStatus,
			"number_process":   t.taskArgs.NumberProcess,
		})
	}()
	var err error
	var reply bool
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

	model.DB().Model(&model.CrontabTask{}).Where("id=?", t.taskArgs.ID).Update(map[string]interface{}{
		"state":          t.taskArgs.State,
		"last_exec_time": t.taskArgs.LastExecTime,
		"number_process": t.taskArgs.NumberProcess,
	})

	if ok := t.waitDependsDone(ctx); !ok {
		cancel()
		errMsg := fmt.Sprintf("[%s %s %s] Execution of dependency script failed\n", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr, t.name)
		t.logContent = append(t.logContent, []byte(errMsg)...)
		writeLog(t.logPath, fmt.Sprintf("%d.log", t.taskArgs.ID), &t.logContent)
		t.taskArgs.LastExitStatus = exitDependError
		isExceptError = true
		if t.taskArgs.UnexpectedExitMail && t.taskArgs.MailTo != "" {
			costTime := time.Now().UnixNano() - start

			err := rpcCall("Logic.SendMail", proto.SendMail{
				MailTo:  strings.Split(t.taskArgs.MailTo, ","),
				Subject: globalConfig.addr + "提醒脚本依赖异常退出",
				Content: fmt.Sprintf(
					"任务名：%s\n详情：%s %v\n开始时间：%s\n耗时：%.4f\n异常：%s",
					t.name, t.taskArgs.Command, t.taskArgs.Args, now.Format("2006-01-02 15:04:05"), float64(costTime)/1000000000, errors.New(errMsg)),
			}, &reply)
			if err != nil {
				log.Println("failed send mail ", err)
			}

		}

	} else {
		// 执行脚本
		if t.taskArgs.Timeout != 0 {
			time.AfterFunc(time.Duration(t.taskArgs.Timeout)*time.Second, func() {
				if flag {
					var reply bool
					isExceptError = true
					switch t.taskArgs.OpTimeout {
					case "email":
						t.taskArgs.LastExitStatus = exitTimeout
						rpcCall("Logic.SendMail", proto.SendMail{
							MailTo:  strings.Split(t.taskArgs.MailTo, ","),
							Subject: globalConfig.addr + "提醒脚本执行超时",
							Content: fmt.Sprintf(
								"任务名：%s\n详情：%s %v\n开始时间：%s\n超时：%ds",
								t.taskArgs.Name, t.taskArgs.Command, t.taskArgs.Args, now.Format("2006-01-02 15:04:05"), t.taskArgs.Timeout),
						}, &reply)
						if err != nil {
							log.Println("failed send mail ", err)
						}

					case "kill":
						t.taskArgs.LastExitStatus = exitTimeout
						cancel()

					case "email_and_kill":
						t.taskArgs.LastExitStatus = exitTimeout
						cancel()
						rpcCall("Logic.SendMail", proto.SendMail{
							MailTo:  strings.Split(t.taskArgs.MailTo, ","),
							Subject: globalConfig.addr + "提醒脚本执行超时",
							Content: fmt.Sprintf(
								"任务名：%s\n详情：%s %v\n开始时间：%s\n超时：%ds",
								t.taskArgs.Name, t.taskArgs.Command, t.taskArgs.Args, now.Format("2006-01-02 15:04:05"), t.taskArgs.Timeout),
						}, &reply)
						if err != nil {
							log.Println("failed send mail ", err)
						}

					case "ignore":
						isExceptError = false
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

		err := wrapExecScript(ctx, fmt.Sprintf("%d.log", t.taskArgs.ID), cmdList, t.logPath, &t.logContent)
		flag = false

		if err != nil {
			if isExceptError == false {
				t.taskArgs.LastExitStatus = exitError
			}

			if t.taskArgs.UnexpectedExitMail {

				err := rpcCall("Logic.SendMail", proto.SendMail{
					MailTo:  strings.Split(t.taskArgs.MailTo, ","),
					Subject: globalConfig.addr + "提醒脚本异常退出",
					Content: fmt.Sprintf(
						"任务名：%s\n详情：%s %v\n开始时间：%s\n异常：%s",
						t.taskArgs.Name, t.taskArgs.Command, t.taskArgs.Args, now.Format("2006-01-02 15:04:05"), err.Error()),
				}, &reply)
				if err != nil {
					log.Println("failed send mail ", err)
				}
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

	if logContent != nil {
		*logContent = t.logContent
	}

	log.Printf("%s:%s %v %s %.3fs %v", t.taskArgs.Name, t.taskArgs.Command, t.taskArgs.Args, t.taskArgs.OpTimeout, float64(t.taskArgs.LastCostTime)/1000000000, err)

}

type handle struct {
	cancel      context.CancelFunc // 取消定时器
	clockChan   chan time.Time
	crontabTask *model.CrontabTask
	taskPool    []*taskEntity
}

type crontab struct {
	taskChan     chan *model.CrontabTask
	stopTaskChan chan *model.CrontabTask
	killTaskChan chan *model.CrontabTask
	handleMap    map[uint]*handle
	lock         sync.RWMutex
}

func newCrontab(taskChanSize int) *crontab {
	return &crontab{
		taskChan:     make(chan *model.CrontabTask, taskChanSize),
		stopTaskChan: make(chan *model.CrontabTask, taskChanSize),
		killTaskChan: make(chan *model.CrontabTask, taskChanSize),
		handleMap:    make(map[uint]*handle),
	}
}

func (c *crontab) add(t *model.CrontabTask) {
	c.taskChan <- t
}

func (c *crontab) quickStart(t *model.CrontabTask, content *[]byte) {
	taskEty := newTaskEntity(t)
	c.lock.Lock()
	if _, ok := c.handleMap[t.ID]; !ok {

		taskPool := make([]*taskEntity, 0)
		c.handleMap[t.ID] = &handle{
			taskPool:    append(taskPool, taskEty),
			crontabTask: t,
		}
		c.lock.Unlock()
	} else {
		c.handleMap[t.ID].taskPool = append(c.handleMap[t.ID].taskPool, taskEty)
		c.lock.Unlock()
	}

	taskEty.exec(content)
}

// stop 停止计划任务并杀死正在执行的脚本进程
func (c *crontab) stop(t *model.CrontabTask) {
	c.kill(t)
	c.stopTaskChan <- t
}

func (c *crontab) update(id uint, fn func(t *model.CrontabTask) error) error {
	var err error
	c.lock.Lock()
	h := c.handleMap[id]
	if h != nil {
		err = fn(h.crontabTask)
	}
	c.lock.Unlock()
	return err
}

// 杀死正在执行的脚本进程
func (c *crontab) kill(t *model.CrontabTask) {
	c.killTaskChan <- t
}

// 删除计划任务
func (c *crontab) delete(t *model.CrontabTask) {
	if model.DB().Unscoped().Delete(t).Error != nil {
		log.Println("failed delete", t.Name, fmt.Sprint("(", t.ID, ")"))
		return
	}

	c.stop(t)
	log.Println("delete", t.Name, t.ID)
}

func (c *crontab) ids() []uint {
	var sli []uint
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
		var crontabTaskList []model.CrontabTask
		model.DB().Model(&model.CrontabTask{}).Update(map[string]interface{}{
			"timer_counter":  0,
			"number_process": 0,
		})
		model.DB().Model(&model.CrontabTask{}).Find(&crontabTaskList)

		for _, v := range crontabTaskList {
			t := v
			if v.State != 0 {
				c.add(&t)
			}
		}

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
					log.Printf("clock:%d is closed", k)
					continue
				}

				// TODO 风险
				select {
				case v.clockChan <- now:
				default:
					// case <-time.After(1 * time.Second):
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
				if h, ok := c.handleMap[t.ID]; !ok {
					ctx, cancel := context.WithCancel(context.Background())
					c.handleMap[t.ID] = &handle{
						cancel:      cancel,
						crontabTask: t,
						clockChan:   make(chan time.Time),
					}
					c.lock.Unlock()
					go c.deal(t, ctx)
					log.Printf("add task %s (ID:%d)", t.Name, t.ID)
				} else {
					if h.cancel == nil {
						ctx, cancel := context.WithCancel(context.Background())
						c.handleMap[t.ID].cancel = cancel
						if c.handleMap[t.ID].clockChan == nil {
							c.handleMap[t.ID].clockChan = make(chan time.Time)
						}
						c.lock.Unlock()
						go c.deal(t, ctx)
					} else {
						c.lock.Unlock()
					}
					log.Printf("task %s (ID:%d) exists", t.Name, t.ID)
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
				if handle, ok := c.handleMap[task.ID]; ok {
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
				if handle, ok := c.handleMap[task.ID]; ok {
					c.lock.Unlock()
					if handle.taskPool != nil {
						for k, v := range handle.taskPool {
							if v.cancel == nil {
								log.Println("kill", task.Name, task.ID, k, "but cancel handler is nul")
							} else {
								v.cancel()
								log.Println("kill", task.Name, task.ID, k)
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

func (c *crontab) deal(task *model.CrontabTask, ctx context.Context) {
	var wgroup sync.WaitGroup
	// 定时计数器用于统计有多少个定时器，当定时器为0时说明没有正在执行的计划
	atomic.AddInt32(&task.TimerCounter, 1)
	task.State = 1
	defer func() {
		atomic.AddInt32(&task.TimerCounter, -1)
		model.DB().Model(&model.CrontabTask{}).Where("id=?", task.ID).Update(map[string]interface{}{
			"state":            0,
			"lat_cost_time":    task.LastCostTime,
			"last_exec_time":   task.LastExecTime,
			"last_exit_status": task.LastExitStatus,
			"number_process":   task.NumberProcess,
			"timer_counter":    task.TimerCounter,
		})
	}()

	model.DB().Model(&model.CrontabTask{}).Where("id=?", task.ID).Update(map[string]interface{}{
		"state":         task.State,
		"timer_counter": task.TimerCounter,
	})

	c.lock.Lock()
	h := c.handleMap[task.ID]
	c.lock.Unlock()
	if h == nil {
		return
	}

	for {

		select {
		case now := <-h.clockChan:

			go func(now time.Time) {
				defer func() {
					if err := recover(); err != nil {
						log.Printf("task panic:%+v\n", *task)
					}
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
			delete(c.handleMap, task.ID)
			c.lock.Unlock()
			log.Printf("stop %s (ID:%d) ok", task.Name, task.ID)
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
		// 并发模式
		syncFlag = pushDepends(t.depends)
	}
	if !syncFlag {
		prefix := fmt.Sprintf("[%s %s] ", time.Now().Format("2006-01-02 15:04:05"), globalConfig.addr)
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

// TDDO select db
func (c *crontab) count() int {
	c.lock.Lock()
	total := len(c.handleMap)
	c.lock.Unlock()
	return total
}
