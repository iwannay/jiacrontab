package admin

import (
	"fmt"
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
		ctx.respBasicError(err)
		return
	}

	ctx.pubEvent(eDesc[reqBody.Action], reqBody.Addr, reqBody)
	ctx.respSucc("", nil)
}

// EditDaemonJob 修改常驻任务，jobID为0时新增
func EditDaemonJob(c iris.Context) {
	var (
		err       error
		reply     int
		ctx       = wrapCtx(c)
		reqBody   EditDaemonJobReqParams
		daemonJob models.DaemonJob
		cla       CustomerClaims
		node      models.Node
	)

	if cla, err = ctx.getClaimsFromToken(); err != nil {
		ctx.respJWTError(err)
		return
	}

	if err = reqBody.verify(ctx); err != nil {
		ctx.respBasicError(err)
		return
	}

	if !node.VerifyUserGroup(cla.UserID, cla.GroupID, reqBody.Addr) {
		ctx.respBasicError(fmt.Errorf("userID:%d groupID:%d permission not allowed", cla.UserID, cla.GroupID))
		return
	}

	daemonJob = models.DaemonJob{
		Name:            reqBody.Name,
		ErrorMailNotify: reqBody.ErrorMailNotify,
		ErrorAPINotify:  reqBody.ErrorMailNotify,
		MailTo:          reqBody.MailTo,
		APITo:           reqBody.APITo,
		UpdatedUserID:   cla.UserID,
		UpdatedUsername: cla.Username,
		Commands:        reqBody.Commands,
		FailRestart:     reqBody.FailRestart,
	}

	daemonJob.ID = reqBody.JobID

	if daemonJob.ID == 0 {
		daemonJob.CreatedUserID = cla.UserID
		daemonJob.CreatedUsername = cla.Username
	}

	if err = rpcCall(reqBody.Addr, "DaemonJob.Edit", daemonJob, &reply); err != nil {
		ctx.respRPCError(err)
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
