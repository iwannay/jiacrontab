package handle

import (
	"github.com/kataras/iris"

	"jiacrontab/libs/rpc"
	"jiacrontab/server/conf"

	"jiacrontab/libs/proto"
	"jiacrontab/model"
	"net/http"
	"strconv"
)

func ListDaemonTask(ctx iris.Context) {

	page := ctx.FormValueDefault("page", "1")
	pagesize := ctx.FormValueDefault("pagesize", "200")
	addr := ctx.FormValue("addr")
	if addr == "" {
		ctx.ViewData("error", "client地址不能为空")
		ctx.View("public/error.html")
		return
	}

	pageInt, err1 := strconv.Atoi(page)
	pagesizeInt, err2 := strconv.Atoi(pagesize)

	if err1 != nil || err2 != nil {
		ctx.ViewData("error", "分页参数错误")
		ctx.View("public/error.html")
		return
	}

	var daemonTaskList []model.DaemonTask

	err := rpc.Call(addr, "DaemonTask.ListDaemonTask", struct{ page, pagesize int }{
		page:     pageInt,
		pagesize: pagesizeInt,
	}, &daemonTaskList)

	if err != nil {
		if err1 != nil || err2 != nil {
			ctx.ViewData("error", err)
			ctx.View("public/error.html")
			return
		}
	}

	ctx.View("daemon/list.html")
}

func EditDaemonTask(ctx iris.Context) {

	var reply bool
	addr := ctx.FormValue("addr")
	taskId := ctx.FormValue("taskId")

	ctx.ViewData("allowCommands", conf.ConfigArgs.AllowCommands)

	if ctx.Request().Method == http.MethodPost {

		name := ctx.PostValueTrim("name")
		mailNotify, err := ctx.PostValueBool("mailNotify")
		mailTo := ctx.PostValueTrim("mailTo")
		command := ctx.PostValue("command")
		args := ctx.PostValue("args")

		if addr == "" || name == "" || command == "" {
			ctx.ViewData("errorMsg", "参数不正确")
			ctx.ViewData("formValues", ctx.FormValues())
			ctx.View("daemon/edit.html")
			return
		}

		remoteArgs := model.DaemonTask{
			Name:       name,
			MailNofity: mailNotify,
			MailTo:     mailTo,
			Command:    command,
			Args:       args,
		}

		err = rpc.Call(addr, "DaemonTask.CreateDaemonTask", remoteArgs, &reply)
		if err != nil {
			ctx.ViewData("formValues", ctx.FormValues())
			ctx.ViewData("errorMsg", err)
			ctx.View("daemon/edit.html")
			return
		}

		ctx.Redirect("/daemon/task/list?addr=" + addr)
	}

	if taskId != "" {

		taskIdInt, err := strconv.Atoi(taskId)
		if err != nil {
			ctx.ViewData("errorMsg", "参数不正确")
			ctx.View("daemon/edit.html")
			return
		}
		var daemonTask model.DaemonTask
		err = rpc.Call(addr, "DaemonTask.GetDaemonTask", taskIdInt, &daemonTask)
		if err != nil {
			ctx.ViewData("errorMsg", "查询不到任务")
			ctx.View("daemon/edit.html")
			return
		}

		ctx.ViewData("daemonTask", daemonTask)
		ctx.ViewData("errorMsg", "参数不正确")
		ctx.View("daemon/edit.html")
		return

	}

	ctx.View("daemon/edit.html")
}

func ActionDaemonTask(ctx iris.Context) {
	var replay bool
	action := ctx.FormValue("action")
	addr := ctx.FormValue("addr")
	taskId, err := strconv.Atoi(ctx.FormValue("taskId"))
	if err != nil {
		ctx.View("public/error.html", map[string]interface{}{
			"error": "invalid taskId",
		})
		return
	}
	var op int
	switch action {
	case "start":
		op = proto.StartDaemonTask
	case "stop":
		op = proto.StopDaemonTask
	case "delete":
		op = proto.StopDaemonTask
	default:
		ctx.View("public/error.html", map[string]interface{}{
			"error": "invalid action",
		})
		return
	}

	err = rpc.Call(addr, "DaemonTask.ActionDaemonTask", proto.ActionDaemonTaskArgs{
		Action: op,
		TaskId: taskId,
	}, &replay)
	if err != nil {
		ctx.View("public/error.html", map[string]interface{}{
			"error": err,
		})
		return
	}

	ctx.Redirect("/daemon/task/list?addr=" + addr)
}
