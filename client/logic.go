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
		for k, _ := range args.Depends {
			args.Depends[k].TaskId = args.Id
		}
		globalStore.Update(func(s *store.Store) {
			s.TaskList[args.Id] = &args
		}).Sync()

		globalCrontab.add(&args)

	} else {
		globalStore.Update(func(s *store.Store) {
			if v, ok2 := s.TaskList[args.Id]; ok2 {
				v.Name = args.Name
				v.Command = args.Command
				v.Args = args.Args
				v.MailTo = args.MailTo
				v.Depends = args.Depends

				for k, _ := range v.Depends {
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
		}).Sync()
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

	if v, ok2 := globalStore.SearchTaskList(args); ok2 {
		globalCrontab.add(v)
		*ok = true
	} else {
		*ok = false
	}

	return nil
}

func (t *Task) Stop(args string, ok *bool) error {

	if v, ok2 := globalStore.SearchTaskList(args); ok2 {
		globalCrontab.stop(v)
		*ok = true
	} else {
		*ok = false
	}

	return nil
}

func (t *Task) Delete(args string, ok *bool) error {

	if v, ok2 := globalStore.SearchTaskList(args); ok2 {
		globalCrontab.delete(v)
		*ok = true
	} else {
		*ok = false
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

		filename = fmt.Sprintf("%s-%s.log", v.Name, v.Id)
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

func (t *Task) SystemInfo(args string, ret *map[string]interface{}) error {
	*ret = libs.SystemInfo(startTime)
	return nil
}

func (t *Task) ResolvedSDepends(args proto.MScript, ok *bool) error {
	defer func() {
		log.Println("exec Task.ResolvedSDepends")
	}()

	if t, ok2 := globalStore.SearchTaskList(args.TaskId); ok2 {
		flag := true
		for k, v := range t.Depends {
			if args.Command+args.Args == v.Command+v.Args {
				t.Depends[k].Done = true
				t.Depends[k].LogContent = args.LogContent
			}

			if t.Depends[k].Done == false {
				flag = false
			}
		}
		if flag {
			var logContent []byte
			for _, v := range t.Depends {
				logContent = append(logContent, v.LogContent...)
			}
			globalCrontab.resolvedDepends(t, logContent)
		}
		*ok = true
	} else {
		*ok = false
	}

	return nil
}

func (t *Task) ExecDepend(args proto.MScript, reply *bool) error {
	globalDepend.Add(args)
	*reply = true
	log.Printf("exec Task.ExecDepend %v", args)
	return nil
}
