package handle

import (
	"strings"

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

	var clientList []model.Client
	model.DB().Find(&clientList)

	if ctx.Request().Method == http.MethodPost {
		if addr == "" {
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

		err := rpc.Call(addr, "DaemonTask.ListDaemonTask", struct{ Page, Pagesize int }{
			Page:     pageInt,
			Pagesize: pagesizeInt,
		}, &daemonTaskList)

		if err != nil {

			ctx.JSON(map[string]interface{}{
				"code": -1,
				"msg":  err.Error(),
			})
			return

		}
		ctx.JSON(map[string]interface{}{
			"code": 0,
			"data": daemonTaskList,
		})
		return
	}

	ctx.ViewData("addrs", clientList)
	ctx.ViewData("addr", addr)
	ctx.View("daemon/list.html")
}

func EditDaemonTask(ctx iris.Context) {
	var daemonTask model.DaemonTask

	addr := ctx.FormValue("addr")
	taskId := ctx.FormValue("taskId")

	ctx.ViewData("allowCommands", conf.AppService.AllowCommands)

	if ctx.Request().Method == http.MethodPost {
		var mailNotify bool
		name := ctx.PostValueTrim("name")

		mailTo := ctx.PostValueTrim("mailTo")
		command := ctx.PostValue("command")
		args := ctx.PostValue("args")

		failedRestart, err := ctx.PostValueBool("failedRestart")
		if addr == "" || name == "" || command == "" || err != nil {
			ctx.ViewData("errorMsg", "参数不正确")
			ctx.ViewData("daemonTask", daemonTask)
			ctx.View("daemon/edit.html")
			return
		}
		if ctx.PostValue("mailNotify") == "true" {
			mailNotify = true
		}

		daemonTask = model.DaemonTask{
			Name:          name,
			MailNotify:    mailNotify,
			MailTo:        mailTo,
			Command:       command,
			FailedRestart: failedRestart,
			Args:          args,
		}

		if taskId != "" {
			v, _ := strconv.Atoi(taskId)
			daemonTask.ID = uint(v)
		}
		var reply int
		err = rpc.Call(addr, "DaemonTask.UpdateDaemonTask", daemonTask, &reply)

		if err != nil {
			ctx.ViewData("daemonTask", daemonTask)
			ctx.ViewData("errorMsg", err)
			ctx.View("daemon/edit.html")
			return
		}

		ctx.Redirect("/daemon/task/list?addr=" + addr)
	}

	if taskId != "" {

		taskIdInt, err := strconv.Atoi(taskId)
		if err != nil {
			ctx.ViewData("errorMsg", err)
			ctx.ViewData("daemonTask", daemonTask)
			ctx.View("daemon/edit.html")
			return
		}

		err = rpc.Call(addr, "DaemonTask.GetDaemonTask", taskIdInt, &daemonTask)
		if err != nil {
			ctx.ViewData("errorMsg", "查询不到任务")
			ctx.ViewData("daemonTask", daemonTask)
			ctx.View("daemon/edit.html")
			return
		}

	}
	ctx.ViewData("daemonTask", daemonTask)
	ctx.View("daemon/edit.html")
}

func ActionDaemonTask(ctx iris.Context) {
	var replay bool
	action := ctx.FormValue("action")
	addr := ctx.FormValue("addr")
	taskIds := ctx.FormValue("taskId")

	var op int
	switch action {
	case "start":
		op = proto.StartDaemonTask
	case "stop":
		op = proto.StopDaemonTask
	case "delete":
		op = proto.DeleteDaemonTask
	default:
		ctx.View("public/error.html", map[string]interface{}{
			"error": "invalid action",
		})
		return
	}

	err := rpc.Call(addr, "DaemonTask.ActionDaemonTask", proto.ActionDaemonTaskArgs{
		Action:  op,
		TaskIds: taskIds,
	}, &replay)
	if err != nil {
		ctx.View("public/error.html", map[string]interface{}{
			"error": err,
		})
		return
	}

	ctx.Redirect("/daemon/task/list?addr=" + addr)
}

func RecentDaemonLog(ctx iris.Context) {
	var r = ctx.Request()
	var searchRet proto.SearchLogResult
	pagesize := 50
	id, err := strconv.Atoi(r.FormValue("taskId"))
	if err != nil {

		ctx.ViewData("error", "参数错误")
		ctx.View("public/error.html")
		return

	}
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil || page == 0 {
		page = 1
	}

	date := r.FormValue("date")
	pattern := r.FormValue("pattern")
	addr := r.FormValue("addr")

	if err := rpc.Call(addr, "DaemonTask.Log", proto.SearchLog{
		TaskId:   id,
		Page:     page,
		Pagesize: pagesize,
		Date:     date,
		Pattern:  pattern,
	}, &searchRet); err != nil {

		ctx.ViewData("error", err)
		ctx.View("public/error.html")
		return

	}
	logList := strings.Split(string(searchRet.Content), "\n")
	ctx.ViewData("logList", logList)
	ctx.ViewData("addr", addr)
	ctx.ViewData("total", searchRet.Total)
	ctx.ViewData("pagesize", pagesize)
	ctx.View("daemon/log.html")

}
