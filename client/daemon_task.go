package main

import (
	"errors"
	"fmt"
	"jiacrontab/libs/finder"
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
			"api_to":         args.ApiTo,
			"mail_notify":    args.MailNotify,
			"api_notify":     args.ApiNotify,
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

func (t *DaemonTask) All(args string, reply *[]model.DaemonTask) error {
	return model.DB().Find(reply).Error
}

func (t *DaemonTask) ActionDaemonTask(args proto.ActionDaemonTaskArgs, reply *bool) error {

	var tasks []model.DaemonTask

	*reply = true

	ret := model.DB().Model(&model.DaemonTask{}).Find(&tasks, "id in(?)", strings.Split(args.TaskIds, ","))

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

func (t *DaemonTask) Log(args proto.SearchLog, reply *proto.SearchLogResult) error {

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

	if args.IsTail {
		fd.SetTail(true)
	}

	rootpath := filepath.Join(globalConfig.logPath, "daemon_task", args.Date)
	err := fd.Search(rootpath, args.Pattern, &reply.Content, args.Page, args.Pagesize)
	reply.Total = int(fd.Count())
	return err

}
