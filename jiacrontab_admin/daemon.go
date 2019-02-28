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
		reqBody GetJobListReqParams
		jobRet  proto.QueryDaemonJobRet
		err     error
	)

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if err = rpcCall(reqBody.Addr, "DaemonJob.List", &proto.QueryJobArgs{
		Page:     reqBody.Page,
		Pagesize: reqBody.Pagesize,
	}, &jobRet); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"list":     jobRet.List,
		"page":     jobRet.Page,
		"pagesize": jobRet.Pagesize,
		"total":    jobRet.Total,
	})
}

func actionDaemonTask(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		err     error
		reply   bool
		ok      bool
		reqBody ActionTaskReqParams
		methods = map[string]int{
			"start":  proto.ActionStartDaemonJob,
			"stop":   proto.ActionStopDaemonJob,
			"delete": proto.ActionDeleteDaemonJob,
		}
		eDesc = map[string]string{
			"start":  event_StartDaemonJob,
			"stop":   event_StopDaemonJob,
			"delete": event_DelDaemonJob,
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

	ctx.pubEvent(eDesc[reqBody.Action], reqBody.Addr, reqBody)
	ctx.respSucc("", nil)
}

func editDaemonJob(c iris.Context) {
	var (
		err       error
		reply     int
		ctx       = wrapCtx(c)
		reqBody   EditDaemonJobReqParams
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
	ctx.pubEvent(event_EditDaemonJob, reqBody.Addr, reqBody)
	ctx.respSucc("", reply)
}

func getDaemonJob(c iris.Context) {
	var (
		ctx       = wrapCtx(c)
		reqBody   GetJobReqParams
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
		reqBody   GetLogReqParams
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
