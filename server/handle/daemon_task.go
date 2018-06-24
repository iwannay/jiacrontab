package handle

import (
	"fmt"

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

	if ctx.Request() == http.MethodPost {
		if addr == "" {
			ctx.ViewData("error", "client地址不能为空")

			ctx.JSON(map[string]interface{}{
				"code": -1,
				"msg":  "addr地址不能为空",
			})
			return
		}

		pageInt, err1 := strconv.Atoi(page)
		pagesizeInt, err2 := strconv.Atoi(pagesize)

		if err1 != nil || err2 != nil {
			ctx.JSON(map[string]interface{}{
				"code": -1,
				"msg":  "分页参数错误",
			})
			return
		}

		var daemonTaskList []model.DaemonTask

		err := rpc.Call(addr, "DaemonTask.ListDaemonTask", struct{ page, pagesize int }{
			page:     pageInt,
			pagesize: pagesizeInt,
		}, &daemonTaskList)

		if err != nil {

			ctx.JSON(map[string]interface{}{
				"code": -1,
				"msg":  err,
			})
			return

		}
		ctx.JSON(map[string]interface{}{
			"code": 0,
			"data": daemonTaskList,
		})
	}

	ctx.View("daemon/list.html")
}

func EditDaemonTask(ctx iris.Context) {

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
		var reply int
		err = rpc.Call(addr, "DaemonTask.CreateDaemonTask", remoteArgs, &reply)
		fmt.Println(reply)
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
