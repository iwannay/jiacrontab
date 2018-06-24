package main

import (
	"jiacrontab/libs/proto"
	"jiacrontab/model"
)

func (t *Task) CreateDaemonTask(args model.DaemonTask, reply *int64) error {

	ret := model.DB().Create(args)

	*reply = ret.RowsAffected
	return ret.Error
}

func (t *Task) ListDaemonTask(args struct{ page, pagesize int }, reply *[]model.DaemonTask) error {
	ret := model.DB().Find(reply).Offset((args.page - 1) * args.pagesize).Limit(args.pagesize).Order("update_at desc")

	return ret.Error
}

func (t *Task) ActionDaemonTask(args proto.ActionDaemonTaskArgs, reply *bool) error {

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

func (t *Task) GetDaemonTask(args int, reply *model.DaemonTask) error {
	ret := model.DB().Find(reply, "task_id", args)
	if (*reply == model.DaemonTask{}) {
		return ret.Error
	}
	return nil
}
