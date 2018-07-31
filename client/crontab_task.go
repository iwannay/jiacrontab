package main

import (
	"errors"
	"fmt"
	"jiacrontab/libs"
	"jiacrontab/libs/finder"
	"jiacrontab/libs/proto"
	"jiacrontab/model"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CrontabTask struct {
}

func (t *CrontabTask) List(args struct{ Page, Pagesize int }, reply *[]model.CrontabTask) error {
	return model.DB().Offset(args.Page - 1).Limit(args.Pagesize).Find(reply).Error
}
func (t *CrontabTask) All(args string, reply *[]model.CrontabTask) error {
	return model.DB().Find(reply).Error
}

func (t *CrontabTask) Update(args model.CrontabTask, ok *bool) error {
	*ok = true
	var err error
	if args.MailTo == "" {
		args.MailTo = globalConfig.mailTo
	}

	if args.ID == 0 {
		ret := model.DB().Create(&args)
		if ret.Error == nil {
			globalCrontab.add(&args)
		}
		err = ret.Error

	} else {

		if args.MaxConcurrent == 0 {
			args.MaxConcurrent = 1
		}
		err = globalCrontab.update(args.ID, func(t *model.CrontabTask) error {
			if t.NumberProcess > 0 {
				return errors.New("can not update when task is running")
			}
			t.Name = args.Name
			t.Command = args.Command
			t.Args = args.Args
			t.MailTo = args.MailTo
			t.ApiTo = args.ApiTo
			t.Depends = args.Depends
			t.UnexpectedExitMail = args.UnexpectedExitMail
			t.UnexpectedExitApi = args.UnexpectedExitApi
			t.PipeCommands = args.PipeCommands
			t.Sync = args.Sync
			t.Timeout = args.Timeout
			t.MaxConcurrent = args.MaxConcurrent
			t.MailTo = args.MailTo
			t.OpTimeout = args.OpTimeout
			t.C = args.C
			return nil

		})

		if err == nil {
			model.DB().Model(&model.CrontabTask{}).Where("id=? and number_process=0", args.ID).Update(map[string]interface{}{
				"name":                 args.Name,
				"command":              args.Command,
				"args":                 args.Args,
				"mail_to":              args.MailTo,
				"depends":              args.Depends,
				"upexpected_exit_mail": args.UnexpectedExitMail,
				"pipe_commands":        args.PipeCommands,
				"sync":                 args.Sync,
				"timeout":              args.Timeout,
				"max_concurrent":       args.MaxConcurrent,
				"op_timeout":           args.OpTimeout,
				"c":                    args.C,
			})

		} else {
			*ok = false
		}

	}

	return err
}
func (t *CrontabTask) Get(args uint, reply *model.CrontabTask) error {
	return model.DB().Find(reply, "id=?", args).Error
}

func (t *CrontabTask) Start(args string, ok *bool) error {
	*ok = true
	ids := strings.Split(args, ",")
	for _, v := range ids {
		var crontabTask model.CrontabTask
		ret := model.DB().Find(&crontabTask, "id=?", libs.ParseInt(v))
		if ret.Error != nil {
			*ok = false
		} else {
			if crontabTask.TimerCounter == 0 {
				globalCrontab.add(&crontabTask)
			}
		}
	}

	return nil
}

func (t *CrontabTask) Stop(args string, ok *bool) error {
	*ok = true
	ids := strings.Split(args, ",")
	for _, v := range ids {
		var crontabTask model.CrontabTask
		ret := model.DB().Find(&crontabTask, "id=?", libs.ParseInt(v))
		if ret.Error != nil {
			*ok = false
		} else {
			globalCrontab.stop(&crontabTask)
		}
	}

	return nil
}

func (t *CrontabTask) StopAll(args []string, ok *bool) error {
	*ok = true
	for _, v := range args {
		var crontabTask model.CrontabTask
		ret := model.DB().Find(&crontabTask, "id", libs.ParseInt(v))
		if ret.Error != nil {
			*ok = false
		} else {
			globalCrontab.stop(&crontabTask)
		}
	}
	return nil
}

func (t *CrontabTask) Delete(args string, ok *bool) error {
	*ok = true
	ids := strings.Split(args, ",")
	for _, v := range ids {
		var crontabTask model.CrontabTask
		ret := model.DB().Find(&crontabTask, "id=?", libs.ParseInt(v))
		if ret.Error != nil {
			*ok = false
		} else {
			globalCrontab.delete(&crontabTask)
		}
	}

	return nil
}

func (t *CrontabTask) Kill(args string, ok *bool) error {

	var crontabTask model.CrontabTask
	ret := model.DB().Find(&crontabTask, "id=?", libs.ParseInt(args))
	if ret.Error != nil {
		*ok = false
	} else {
		globalCrontab.kill(&crontabTask)
		*ok = true
	}

	return nil
}

func (t *CrontabTask) QuickStart(args string, reply *[]byte) error {

	var crontabTask model.CrontabTask
	ret := model.DB().Find(&crontabTask, "id=?", libs.ParseInt(args))

	if ret.Error == nil {
		globalCrontab.quickStart(&crontabTask, reply)
	} else {
		*reply = []byte("failed to start")
	}
	return nil

}

func (t *CrontabTask) Log(args proto.SearchLog, reply *proto.SearchLogResult) error {

	fd := finder.NewFinder(1000000, func(info os.FileInfo) bool {
		basename := filepath.Base(info.Name())
		arr := strings.Split(basename, ".")
		if len(arr) != 2 {
			return false
		}
		if arr[1] == "log" && arr[0] == fmt.Sprint(args.TaskId) {
			return true
		}
		return false
	})

	if args.Date == "" {
		args.Date = time.Now().Format("2006/01/02")
	}

	rootpath := filepath.Join(globalConfig.logPath, "crontab_task", args.Date)
	err := fd.Search(rootpath, args.Pattern, &reply.Content, args.Page, args.Pagesize)
	reply.Total = int(fd.Count())
	return err

}

func (t *CrontabTask) ResolvedDepends(args model.DependsTask, reply *bool) error {
	var err error
	if args.Err != "" {
		err = errors.New(args.Err)
	}

	globalCrontab.lock.Lock()
	if h, ok := globalCrontab.handleMap[args.TaskId]; ok {
		globalCrontab.lock.Unlock()

		for _, v := range h.taskPool {
			if v.id == args.TaskEntityId {
				for _, v2 := range v.depends {
					if v2.id == args.Id {
						v2.dest = args.Dest
						v2.from = args.From
						v2.logContent = args.LogContent
						v2.err = err
						v2.done = true
						*reply = filterDepend(v2)
						return nil
					}
				}
			}
		}
	} else {
		globalCrontab.lock.Unlock()
	}

	log.Printf("resolvedDepends not exists taskId:%d", args.TaskId)

	*reply = false
	return nil
}

func (t *CrontabTask) ExecDepend(args model.DependsTask, reply *bool) error {
	globalDepend.Add(&dependScript{
		id:           args.Id,
		taskEntityId: args.TaskEntityId,
		taskId:       args.TaskId,
		dest:         args.Dest,
		from:         args.From,
		name:         args.Name,
		command:      args.Command,
		args:         args.Args,
	})
	*reply = true
	log.Printf("task %s %s %s add to execution queue ", args.Name, args.Command, args.Args)
	return nil
}

func (t *CrontabTask) Ping(args *proto.EmptyArgs, reply *proto.EmptyReply) error {
	return nil
}
