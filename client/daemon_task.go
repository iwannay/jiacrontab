package main

import (
	"jiacrontab/libs/proto"
	"jiacrontab/model"
)

type DaemonTask struct {
}

func (t *DaemonTask) CreateDaemonTask(args model.DaemonTask, reply *int64) error {

	ret := model.DB().Create(&args)

	*reply = ret.RowsAffected
	return ret.Error
}

func (t *DaemonTask) ListDaemonTask(args struct{ Page, Pagesize int }, reply *[]model.DaemonTask) error {
	ret := model.DB().Find(reply).Offset((args.Page - 1) * args.Pagesize).Limit(args.pagesize).Order("update_at desc")

	return ret.Error
}

func (t *DaemonTask) ActionDaemonTask(args proto.ActionDaemonTaskArgs, reply *bool) error {

	var task model.DaemonTask

	*reply = false

	ret := model.DB().Find(&task, "task_id=?", args.TaskId)

	if (task == model.DaemonTask{}) {

		return ret.Error
	}

	globalDaemon.add(&daemonTask{
		task:   &task,
		action: args.Action,
	})
	return nil
}

func (t *DaemonTask) GetDaemonTask(args int, reply *model.DaemonTask) error {
	ret := model.DB().Find(reply, "task_id", args)
	if (*reply == model.DaemonTask{}) {
		return ret.Error
	}
	return nil
}
