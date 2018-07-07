package main

import (
	"errors"
	"fmt"
	"jiacrontab/libs"
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
	if args.MailTo == "" {
		args.MailTo = globalConfig.mailTo
	}

	if args.ID == 0 {
		ret := model.DB().Create(&args)
		if ret.Error == nil {

			globalCrontab.add(&args)
		}

	} else {
		var crontabTask model.CrontabTask
		ret := model.DB().Find(&crontabTask, "id=?", args.ID)
		if ret.Error == nil {
			if crontabTask.NumberProcess > 0 {
				return errors.New("can not update when task is running")
			}

			crontabTask.Name = args.Name
			crontabTask.Command = args.Command
			crontabTask.Args = args.Args
			crontabTask.MailTo = args.MailTo
			crontabTask.Depends = args.Depends
			crontabTask.UnexpectedExitMail = args.UnexpectedExitMail
			crontabTask.PipeCommands = args.PipeCommands
			crontabTask.Sync = args.Sync
			crontabTask.Timeout = args.Timeout
			crontabTask.MaxConcurrent = args.MaxConcurrent
			if crontabTask.MaxConcurrent == 0 {
				crontabTask.MaxConcurrent = 1
			}

			crontabTask.MailTo = args.MailTo
			crontabTask.OpTimeout = args.OpTimeout
			crontabTask.C = args.C
			model.DB().Debug().Model(&model.CrontabTask{}).Where("id=? and number_process=0", crontabTask.ID).Update(map[string]interface{}{
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

	return nil
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
		fmt.Printf("id", libs.ParseInt(v))
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

func (t *CrontabTask) Log(args string, reply *[]byte) error {
	staticDir := filepath.Join(globalConfig.logPath, "crontab_task", time.Now().Format("2006/01"))
	var filename string
	var crontabTask model.CrontabTask

	ret := model.DB().Find(&crontabTask, "id=?", args)

	if ret.Error == nil {
		filename = fmt.Sprintf("%d.log", crontabTask.ID)
	}

	if filename == "" {
		return errors.New("log file not found")
	}
	fp := filepath.Join(staticDir, filename)
	f, err := os.Open(fp)
	defer f.Close()
	if err != nil {
		return err
	}
	fStat, err := f.Stat()
	if err != nil {
		return err
	}
	limit := int64(1024 * 1024)
	var offset int64
	var buffer []byte
	if fStat.Size() > limit {
		buffer = make([]byte, limit)
		offset = fStat.Size() - limit
	} else {
		offset = 0
		buffer = make([]byte, fStat.Size())
	}
	f.Seek(offset, os.SEEK_CUR)

	_, err = f.Read(buffer)
	*reply = buffer

	return err
}

func (t *CrontabTask) ResolvedDepends(args model.DependsTask, reply *bool) error {

	var err error
	if args.Err != "" {
		err = errors.New(args.Err)
	}

	idArr := strings.Split(args.TaskId, "-")
	globalCrontab.lock.Lock()
	i := uint(libs.ParseInt(idArr[0]))
	if h, ok := globalCrontab.handleMap[i]; ok {
		globalCrontab.lock.Unlock()
		for _, v := range h.taskPool {
			if v.id == idArr[1] {
				for _, v2 := range v.depends {
					if v2.id == args.TaskId {
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

	log.Printf("resolvedDepends: %s is not exists", args.Name)

	*reply = false
	return nil
}

func (t *CrontabTask) ExecDepend(args model.DependsTask, reply *bool) error {

	globalDepend.Add(&dependScript{
		id:      args.TaskId,
		dest:    args.Dest,
		from:    args.From,
		name:    args.Name,
		command: args.Command,
		args:    args.Args,
	})
	*reply = true
	log.Printf("task %s %s %s add to execution queue ", args.Name, args.Command, args.Args)
	return nil
}
