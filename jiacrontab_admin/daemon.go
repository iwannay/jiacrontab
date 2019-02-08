package admin

import (
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"strings"

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
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if err = rpcCall(reqBody.Addr, "DaemonJob.GetDaemonJobList", struct{ Page, Pagesize int }{
		Page:     reqBody.Page,
		Pagesize: reqBody.Pagesize,
	}, &jobList); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", jobList)
}

func actionDaemonTask(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		err     error
		reply   bool
		ok      bool
		reqBody actionTaskReqParams
		methods = map[string]int{
			"start":  proto.ActionStartDaemonTask,
			"stop":   proto.ActionStopDaemonTask,
			"delete": proto.ActionDeleteDaemonTask,
		}
		action int
	)

	if action, ok = methods[reqBody.Action]; !ok {
		ctx.respError(proto.Code_Error, "参数错误", nil)
		return
	}
	if err = rpcCall(reqBody.Addr, "DaemonJob.ActionDaemonTask", proto.ActionDaemonJobArgs{
		Action: action,
		JobIDs: reqBody.JobIDs,
	}, &reply); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", nil)
}

func editDaemonJob(c iris.Context) {
	var (
		err       error
		reply     int
		ctx       = wrapCtx(c)
		reqBody   editDaemonJobReqParams
		daemonJob models.DaemonJob
	)

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	daemonJob = models.DaemonJob{
		Name:            reqBody.Name,
		ErrorMailNotify: reqBody.ErrorMailNotify,
		ErrorAPINotify:  reqBody.ErrorMailNotify,
		MailTo:          reqBody.MailTo,
		APITo:           reqBody.APITo,
		Commands:        reqBody.Commands,
		FailRestart:     reqBody.FailRestart,
	}

	if err = rpcCall(reqBody.Addr, "DaemonJob.EditDaemonJob", daemonJob, &reply); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", reply)
}

func getDaemonJob(c iris.Context) {
	var (
		ctx       = wrapCtx(c)
		reqBody   getJobReqParams
		daemonJob models.DaemonJob
		err       error
	)
	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if err = rpcCall(reqBody.Addr, "daemonJob.GetDaemonJob", reqBody.JobID, &daemonJob); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", daemonJob)
}

func getRecentDaemonLog(c iris.Context) {
	var (
		err       error
		ctx       = wrapCtx(c)
		searchRet proto.SearchLogResult
		reqBody   getLogReqParams
		logList   []string
	)

	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if err := rpc.Call(reqBody.Addr, "DaemonJob.Log", proto.SearchLog{
		JobID:    reqBody.JobID,
		Page:     reqBody.Page,
		Pagesize: reqBody.Pagesize,
		Date:     reqBody.Date,
		Pattern:  reqBody.Pattern,
		IsTail:   reqBody.IsTail,
	}, &searchRet); err != nil {
		goto failed
	}

	logList = strings.Split(string(searchRet.Content), "\n")

	ctx.respSucc("", map[string]interface{}{
		"logList":  logList,
		"curAddr":  reqBody.Addr,
		"total":    searchRet.Total,
		"page":     reqBody.Page,
		"pagesize": reqBody.Pagesize,
	})
	return

failed:
	ctx.respError(proto.Code_Error, "", err.Error())
}
