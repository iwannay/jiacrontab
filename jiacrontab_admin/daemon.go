package admin

import (
	"errors"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"strings"
)

func GetDaemonJobList(ctx *myctx) {
	var (
		reqBody GetJobListReqParams
		jobRet  proto.QueryDaemonJobRet
		err     error
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if err = rpcCall(reqBody.Addr, "DaemonJob.List", &proto.QueryJobArgs{
		Page:      reqBody.Page,
		Pagesize:  reqBody.Pagesize,
		SearchTxt: reqBody.SearchTxt,
		Root:      ctx.claims.Root,
		UserID:    ctx.claims.UserID,
	}, &jobRet); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"list":     jobRet.List,
		"page":     jobRet.Page,
		"pagesize": jobRet.Pagesize,
		"total":    jobRet.Total,
	})
}

func ActionDaemonTask(ctx *myctx) {
	var (
		err     error
		reply   []models.DaemonJob
		ok      bool
		reqBody ActionTaskReqParams

		methods = map[string]string{
			"start":  "DaemonJob.Start",
			"delete": "DaemonJob.Delete",
			"stop":   "DaemonJob.Stop",
		}

		eDesc = map[string]string{
			"start":  event_StartDaemonJob,
			"delete": event_DelDaemonJob,
			"stop":   event_StopDaemonJob,
		}
		method string
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respBasicError(err)
	}

	if method, ok = methods[reqBody.Action]; !ok {
		ctx.respBasicError(errors.New("参数错误"))
		return
	}

	if err = rpcCall(reqBody.Addr, method, proto.ActionJobsArgs{
		UserID: ctx.claims.UserID,
		JobIDs: reqBody.JobIDs,
		Root:   ctx.claims.Root,
	}, &reply); err != nil {
		ctx.respRPCError(err)
		return
	}

	if len(reply) > 0 {
		var targetNames []string
		for _, v := range reply {
			targetNames = append(targetNames, v.Name)
		}
		ctx.pubEvent(strings.Join(targetNames, ","), eDesc[reqBody.Action], models.EventSourceName(reqBody.Addr), reqBody)
	}

	ctx.respSucc("", nil)
}

// EditDaemonJob 修改常驻任务，jobID为0时新增
func EditDaemonJob(ctx *myctx) {
	var (
		err       error
		reply     models.DaemonJob
		reqBody   EditDaemonJobReqParams
		daemonJob models.DaemonJob
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
	}

	daemonJob = models.DaemonJob{
		Name:            reqBody.Name,
		GroupID:         ctx.claims.GroupID,
		ErrorMailNotify: reqBody.ErrorMailNotify,
		ErrorAPINotify:  reqBody.ErrorAPINotify,
		MailTo:          reqBody.MailTo,
		APITo:           reqBody.APITo,
		UpdatedUserID:   ctx.claims.UserID,
		UpdatedUsername: ctx.claims.Username,
		Command:         reqBody.Command,
		WorkDir:         reqBody.WorkDir,
		WorkEnv:         reqBody.WorkEnv,
		Code:            reqBody.Code,
		RetryNum:        reqBody.RetryNum,
		FailRestart:     reqBody.FailRestart,
		Status:          models.StatusJobUnaudited,
	}

	daemonJob.ID = reqBody.JobID

	if daemonJob.ID == 0 {
		daemonJob.CreatedUserID = ctx.claims.UserID
		daemonJob.CreatedUsername = ctx.claims.Username
	}
	if ctx.claims.Root || ctx.claims.GroupID == models.SuperGroup.ID {
		daemonJob.Status = models.StatusJobOk
	} else {
		daemonJob.Status = models.StatusJobUnaudited
	}

	if err = rpcCall(reqBody.Addr, "DaemonJob.Edit", proto.EditDaemonJobArgs{
		Job:  daemonJob,
		Root: ctx.claims.Root,
	}, &reply); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.pubEvent(reply.Name, event_EditDaemonJob, models.EventSourceName(reqBody.Addr), reqBody)
	ctx.respSucc("", reply)
}

func GetDaemonJob(ctx *myctx) {
	var (
		reqBody   GetJobReqParams
		daemonJob models.DaemonJob
		err       error
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
	}

	if err = rpcCall(reqBody.Addr, "DaemonJob.Get", proto.GetJobArgs{
		UserID: ctx.claims.UserID,
		Root:   ctx.claims.Root,
		JobID:  reqBody.JobID,
	}, &daemonJob); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.respSucc("", daemonJob)
}

func GetRecentDaemonLog(ctx *myctx) {
	var (
		err       error
		searchRet proto.SearchLogResult
		reqBody   GetLogReqParams
		logList   []string
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if err := rpc.Call(reqBody.Addr, "DaemonJob.Log", proto.SearchLog{
		JobID:    reqBody.JobID,
		Offset:   reqBody.Offset,
		Pagesize: reqBody.Pagesize,
		Date:     reqBody.Date,
		Pattern:  reqBody.Pattern,
		IsTail:   reqBody.IsTail,
	}, &searchRet); err != nil {
		ctx.respRPCError(err)
		return
	}

	logList = strings.Split(string(searchRet.Content), "\n")

	ctx.respSucc("", map[string]interface{}{
		"logList":  logList,
		"curAddr":  reqBody.Addr,
		"offset":   searchRet.Offset,
		"filesize": searchRet.FileSize,
		"pagesize": reqBody.Pagesize,
	})
}
