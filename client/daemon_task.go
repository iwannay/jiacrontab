package main

import (
	"errors"
	"fmt"
	"jiacrontab/libs/proto"
	"jiacrontab/model"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DaemonTask struct {
}

func (t *DaemonTask) UpdateDaemonTask(args model.DaemonTask, reply *int64) error {

	if args.ID != 0 {
		var daemonTask model.DaemonTask
		ret := model.DB().Find(&daemonTask, "id=?", args.ID)
		if daemonTask.ProcessNum != 0 {
			return errors.New("can not update when task is running")
		}
		ret = model.DB().Model(&model.DaemonTask{}).Where("id=?", args.ID).Update(map[string]interface{}{
			"name":           args.Name,
			"mail_to":        args.MailTo,
			"mail_notify":    args.MailNotify,
			"failed_restart": args.FailedRestart,
			"command":        args.Command,
			"args":           args.Args,
			"status":         args.Status,
		})
		*reply = ret.RowsAffected
		return ret.Error
	}

	ret := model.DB().Create(&args)
	*reply = ret.RowsAffected
	return ret.Error
}

func (t *DaemonTask) ListDaemonTask(args struct{ Page, Pagesize int }, reply *[]model.DaemonTask) error {
	ret := model.DB().Find(reply).Offset((args.Page - 1) * args.Pagesize).Limit(args.Pagesize).Order("update_at desc")

	return ret.Error
}

func (t *DaemonTask) ActionDaemonTask(args proto.ActionDaemonTaskArgs, reply *bool) error {

	var tasks []model.DaemonTask

	*reply = false

	ret := model.DB().Debug().Model(&model.DaemonTask{}).Find(&tasks, "id in(?)", strings.Split(args.TaskIds, ","))

	if ret.Error != nil {
		return ret.Error
	}

	for _, v := range tasks {
		task := v
		globalDaemon.add(&daemonTask{
			task:   &task,
			action: args.Action,
		})
	}

	return nil
}

func (t *DaemonTask) GetDaemonTask(args int, reply *model.DaemonTask) error {
	ret := model.DB().Find(reply, "id=?", args)
	if (*reply == model.DaemonTask{}) {
		return ret.Error
	}
	return nil
}

func (t *DaemonTask) Log(args int, ret *[]byte) error {
	fp := filepath.Join(globalConfig.logPath, "daemon_task", time.Now().Format("2006/01"), fmt.Sprint(args, ".log"))
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
