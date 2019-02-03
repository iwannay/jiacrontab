package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/proto"

	"github.com/kataras/iris"
)

func getDaemonJobList(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		reqBody jobListReqParams
		jobList []models.DaemonJob
		err     error
	)

	if err = reqBody.verify(ctx); err != nil {
		return ctx.respError(proto.Code_Error, err.Error(), nil)
	}

	if err = rpcCall(reqBody.Addr, "DaemonTask.GetDaemonJobList", {
		Page:     reqBody.Page,
		Pagesize: reqBody.Pagesize,
	},&jobList); err != nil {
		return ctx.respError(proto.Code_Error, err.Error(), nil)
	}

	return ctx.respSucc("", jobList)
}

func actionDaemonTask(c iris.Context) {
	var (
		ctx = wrapCtx(c)
		err error
		reply bool
		ok bool
		reqBody actionTaskReqParams
		methods = map[string]string{
			"start":proto.StartDaemonTask,
			"stop":proto.StopDaemonTask,
			"delete":proto.DeleteDaemonTask,
		}
		action string
	)

	if action,ok = methods[reqBody.Action]; !ok {
		return ctx.respError(proto.Code_Error, "参数错误")
	}
	if err = rpcCall(addr, "DaemonTask.ActionDaemonTask",proto.ActionDaemonTaskArgs{
		Action:  op,
		TaskIds: taskIds,
	}, &reply); err != nil {
		return ctx.respError(proto.Code_Error, err.Error())
	}

	return ctx.respSucc("", nil)

}

func editDaemonJob(c iris.Context) {
	var (
		err error
		reply int
		ctx = wrapCtx(c)
		reqBody editDaemonJobReqParams
		daemonJob models.DaemonJob
	)

	if err = reqBody.verify(ctx);err != nil {
		goto failed
	}

	daemonJob = models.DaemonJob{
		Name:reqBody.Name,
		ErrorMailNotify:reqBody.ErrorMailNotify,
		ErrorAPINotify:reqBody.ErrorMailNotify,
		MailTo:reqBody.MailTo,
		APITo:reqBody.APITo,
		Commands:reqBody.Commands,
		FailRestart:reqBody.FailRestart,
	}

	if err = rpcCall(addr, "DaemonTask.EditDaemonJob", daemonJob, &reply); err != nil {
		return ctx.respError(proto.Code_Error, err.Error(), nil)
	}

	return ctx.respSucc("", reply)
}

func getDaemonJob(c iris.Context) {
	var (
		ctx = wrapCtx(c)
		reqBody getJobReqParams
		daemonJob models.DaemonJob
		err error
	)
	if err = reqBody.verify(ctx); err != nil {
		return ctx.respError(proto.Code_Error, err.Error(), nil)
	}

	if err = rpcCall(addr, "daemonJob.GetDaemonJob", reqBody.JobID, &daemonJob); err != nil {
		return ctx.respError(proto.Code_Error, err.Error(), nil)
	}

	return ctx.respSucc("", daemonJob)
}

func getRecentDaemonLog(c iris.Context) {
	var (
		err       error
		ctx       = wrapCtx(c)
		searchRet proto.SearchLogResult
		reqBody   getLogReqParams
	)


	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if err := rpc.Call(reqBody.Addr, "DaemonTask.Log", proto.SearchLog{
		JobID:    reqBody.JobID,
		Page:     reqBody.Page,
		Pagesize: reqBody.Pagesize,
		Date:     reqBody.Date,
		Pattern:  reqBody.pattern,
		IsTail:   isTail,
	}, &searchRet); err != nil {
		ctx.ViewData("error", err)
	}

	logList := strings.Split(string(searchRet.Content), "\n")

	return ctx.respSucc("", map[string]interface{}{
		"logList":  logList,
		"curAddr":  reqBody.Addr,
		"total":    searchRet.Total,
		"page":     reqBody.Page,
		"pagesize": reqBody.Pagesize,
	})

failed:
	return ctx.respError(proto.Code_Error, "", err.Error())
}