package admin

import (
	"errors"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"strings"

	"github.com/kataras/iris"
)

func GetDaemonJobList(c iris.Context) {
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

func ActionDaemonTask(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
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

	if err = reqBody.verify(ctx); err != nil {
		ctx.respBasicError(err)
	}

	if method, ok = methods[reqBody.Action]; !ok {
		ctx.respBasicError(errors.New("参数错误"))
		return
	}

	if err = rpcCall(reqBody.Addr, method, reqBody.JobIDs, &reply); err != nil {
		ctx.respRPCError(err)
		return
	}

	var targetNames []string
	for _, v := range reply {
		targetNames = append(targetNames, v.Name)
	}

	ctx.pubEvent(strings.Join(targetNames, ","), eDesc[reqBody.Action], reqBody.Addr, reqBody)
	ctx.respSucc("", nil)
}

// EditDaemonJob 修改常驻任务，jobID为0时新增
func EditDaemonJob(c iris.Context) {
	var (
		err       error
		reply     models.DaemonJob
		ctx       = wrapCtx(c)
		reqBody   EditDaemonJobReqParams
		daemonJob models.DaemonJob
	)

	if err = ctx.parseClaimsFromToken(); err != nil {
		ctx.respJWTError(err)
		return
	}

	if err = reqBody.verify(ctx); err != nil {
		ctx.respBasicError(err)
		return
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
	}

	daemonJob = models.DaemonJob{
		Name:            reqBody.Name,
		ErrorMailNotify: reqBody.ErrorMailNotify,
		ErrorAPINotify:  reqBody.ErrorMailNotify,
		MailTo:          reqBody.MailTo,
		APITo:           reqBody.APITo,
		UpdatedUserID:   ctx.claims.UserID,
		UpdatedUsername: ctx.claims.Username,
		Commands:        reqBody.Commands,
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

	if err = rpcCall(reqBody.Addr, "DaemonJob.Edit", daemonJob, &reply); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.pubEvent(reply.Name, event_EditDaemonJob, reqBody.Addr, reqBody)
	ctx.respSucc("", reply)
}

func GetDaemonJob(c iris.Context) {
	var (
		ctx       = wrapCtx(c)
		reqBody   GetJobReqParams
		daemonJob models.DaemonJob
		err       error
	)

	if err = ctx.parseClaimsFromToken(); err != nil {
		ctx.respJWTError(err)
		return
	}

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
	}

	if err = rpcCall(reqBody.Addr, "DaemonJob.Get", reqBody.JobID, &daemonJob); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", daemonJob)
}

func GetRecentDaemonLog(c iris.Context) {
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
