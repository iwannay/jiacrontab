package main

import (
	"errors"
	"fmt"
	"jiacrontab/client/store"
	"jiacrontab/libs"
	"jiacrontab/libs/proto"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Task struct {
}

func (t *Task) List(args struct{ page, pagesize int }, reply *[]*proto.TaskArgs) error {
	if taskList, ok := globalStore.GetTaskList(); ok {
		for _, v := range taskList {
			*reply = append(*reply, v)
		}
	}
	return nil
}

func (t *Task) All(args string, reply *proto.Mdata) error {
	if taskList, ok := globalStore.GetTaskList(); ok {
		*reply = taskList
	}

	return nil

}

func (t *Task) Update(args proto.TaskArgs, ok *bool) error {
	*ok = true
	if args.MailTo == "" {
		args.MailTo = globalConfig.mailTo
	}

	if args.Id == "" {

		args.Id = strconv.Itoa(int(libs.RandNum()))
		for k := range args.Depends {
			args.Depends[k].TaskId = args.Id
		}
		globalStore.Update(func(s *store.Store) {
			s.TaskList[args.Id] = &args
		}).Sync()

		globalCrontab.add(&args)

	} else {
		if v, ok2 := globalStore.SearchTaskList(args.Id); ok2 {

			if v.NumberProcess > 0 {
				return errors.New("can not update when task is running")
			}

			v.Name = args.Name
			v.Command = args.Command
			v.Args = args.Args
			v.MailTo = args.MailTo
			v.Depends = args.Depends
			v.UnexpectedExitMail = args.UnexpectedExitMail
			v.PipeCommands = args.PipeCommands
			v.Sync = args.Sync

			for k := range v.Depends {
				v.Depends[k].TaskId = args.Id

			}
			v.Timeout = args.Timeout
			v.MaxConcurrent = args.MaxConcurrent
			if v.MaxConcurrent == 0 {
				v.MaxConcurrent = 1
			}

			v.MailTo = args.MailTo
			v.OpTimeout = args.OpTimeout
			v.C = args.C
		} else {
			*ok = false
		}
		globalStore.Sync()

	}

	return nil
}
func (t *Task) Get(args string, task *proto.TaskArgs) error {

	if v, ok := globalStore.SearchTaskList(args); ok {
		*task = *v
	}

	return nil
}

func (t *Task) Start(args string, ok *bool) error {
	*ok = true
	ids := strings.Split(args, ",")
	for _, v := range ids {
		if val, ok2 := globalStore.SearchTaskList(v); ok2 {
			if val.TimerCounter == 0 {
				globalCrontab.add(val)
			}
		} else {
			*ok = false
		}
	}

	return nil
}

func (t *Task) Stop(args string, ok *bool) error {
	*ok = true
	ids := strings.Split(args, ",")
	for _, v := range ids {
		if val, ok2 := globalStore.SearchTaskList(v); ok2 {
			globalCrontab.stop(val)
		} else {
			*ok = false
		}
	}

	return nil
}

func (t *Task) StopAll(args []string, ok *bool) error {
	*ok = true
	for _, v := range args {
		if val, ok2 := globalStore.SearchTaskList(v); ok2 {
			globalCrontab.stop(val)
		} else {
			*ok = false
		}
	}
	return nil
}

func (t *Task) Delete(args string, ok *bool) error {
	*ok = true
	ids := strings.Split(args, ",")
	for _, v := range ids {
		if val, ok2 := globalStore.SearchTaskList(v); ok2 {
			globalCrontab.delete(val)

		} else {
			*ok = false
		}
	}

	return nil
}

func (t *Task) Kill(args string, ok *bool) error {
	if v, ok2 := globalStore.SearchTaskList(args); ok2 {
		globalCrontab.kill(v)
		*ok = true
	} else {
		*ok = false
	}

	return nil
}

func (t *Task) QuickStart(args string, ret *[]byte) error {

	if v, ok := globalStore.SearchTaskList(args); ok {
		globalCrontab.quickStart(v, ret)
	} else {
		*ret = []byte("failed to start")
	}
	return nil

}

func (t *Task) Log(args string, ret *[]byte) error {
	staticDir := filepath.Join(globalConfig.logPath, strconv.Itoa(time.Now().Year()), time.Now().Month().String())
	var filename string

	if v, ok := globalStore.SearchTaskList(args); ok {

		filename = fmt.Sprintf("%s.log", v.Name)
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
	*ret = buffer

	return err
}

func (t *Task) ResolvedDepends(args proto.MScript, ok *bool) error {

	var err error
	if args.Err != "" {
		err = errors.New(args.Err)
	}

	idArr := strings.Split(args.TaskId, "-")
	globalCrontab.lock.Lock()
	if h, ok2 := globalCrontab.handleMap[idArr[0]]; ok2 {
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
						*ok = filterDepend(v2)
						return nil
					}
				}
			}
		}
	} else {
		globalCrontab.lock.Unlock()
	}

	log.Printf("resolvedDepends: %s is not exists", args.Name)

	*ok = false
	return nil
}

func (t *Task) ExecDepend(args proto.MScript, reply *bool) error {

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
