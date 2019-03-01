package admin

import (
	"errors"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"strings"

	"github.com/kataras/iris"
)

func getJobList(c iris.Context) {

	var (
		jobRet       proto.QueryCrontabJobRet
		ctx          = wrapCtx(c)
		err          error
		reqBody      GetJobListReqParams
		rpcReqParams proto.QueryJobArgs
	)

	if err = reqBody.verify(ctx); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	rpcReqParams.Page = reqBody.Page
	rpcReqParams.Pagesize = reqBody.Pagesize

	if err := rpc.Call(reqBody.Addr, "CrontabJob.List", rpcReqParams, &jobRet); err != nil {
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

func getRecentLog(c iris.Context) {
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

	if err = rpc.Call(reqBody.Addr, "CrontabJob.Log", proto.SearchLog{
		JobID:    reqBody.JobID,
		Page:     reqBody.Page,
		Pagesize: reqBody.Pagesize,
		Date:     reqBody.Date,
		Pattern:  reqBody.Pattern,
		IsTail:   reqBody.IsTail,
	}, &searchRet); err != nil {
		ctx.ViewData("error", err)
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

func editJob(c iris.Context) {
	var (
		err     error
		reply   int
		ctx     = wrapCtx(c)
		reqBody EditJobReqParams
		rpcArgs models.CrontabJob
		userID  uint
		cla     CustomerClaims
		node    models.Node
	)

	cla, _ = ctx.getClaimsFromToken()

	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if !node.VerifyUserGroup(cla.UserID, cla.GroupID, reqBody.Addr) {
		err = errors.New("permission not allowed")
		goto failed
	}
	if reqBody.ID != 0 {
		// 修改job时要判断权限
		if !(cla.GroupID == 0 || cla.Root) {
			userID = cla.UserID
		}
	} else {
		userID = cla.UserID
	}

	rpcArgs = models.CrontabJob{
		Name:     reqBody.Name,
		Commands: reqBody.Commands,
		TimeArgs: models.TimeArgs{
			Month:   reqBody.Month,
			Day:     reqBody.Day,
			Hour:    reqBody.Hour,
			Minute:  reqBody.Minute,
			Weekday: reqBody.Weekday,
			Second:  reqBody.Second,
		},
		UserID:          userID,
		WorkDir:         reqBody.WorkDir,
		WorkUser:        reqBody.WorkUser,
		WorkEnv:         reqBody.WorkEnv,
		RetryNum:        reqBody.RetryNum,
		Timeout:         reqBody.Timeout,
		TimeoutTrigger:  reqBody.TimeoutTrigger,
		MailTo:          reqBody.MailTo,
		APITo:           reqBody.APITo,
		MaxConcurrent:   reqBody.MaxConcurrent,
		DependJobs:      reqBody.DependJobs,
		ErrorMailNotify: reqBody.ErrorMailNotify,
		ErrorAPINotify:  reqBody.ErrorAPINotify,
		IsSync:          reqBody.IsSync,
	}

	rpcArgs.ID = reqBody.ID

	if err = rpcCall(reqBody.Addr, "CrontabJob.Edit", rpcArgs, &reply); err != nil {
		goto failed
	}
	ctx.pubEvent(event_EditCronJob, reqBody.Addr, reqBody)
	ctx.respSucc("", reply)
	return

failed:
	ctx.respError(proto.Code_Error, err.Error(), nil)
}

func actionTask(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		err     error
		reply   bool
		ok      bool
		method  string
		reqBody ActionTaskReqParams
		methods = map[string]string{
			"start":  "CrontabJob.Start",
			"stop":   "CrontabJob.Stop",
			"delete": "CrontabJob.Delete",
			"kill":   "CrontabJob.Kill",
		}

		eDesc = map[string]string{
			"start":  event_StartCronJob,
			"stop":   event_StopCronJob,
			"delete": event_DelCronJob,
			"kill":   event_KillCronJob,
		}
	)

	if reqBody.verify(ctx); err != nil {
		goto failed
	}

	if method, ok = methods[reqBody.Action]; !ok {
		err = errors.New("action无法识别")
		goto failed
	}

	if err = rpcCall(reqBody.Addr, method, reqBody.JobIDs, &reply); err != nil {
		goto failed
	}

	ctx.pubEvent(eDesc[reqBody.Action], reqBody.Addr, reqBody)
	ctx.respSucc("", reply)
	return

failed:
	ctx.respError(proto.Code_Error, err.Error(), nil)

}

func execTask(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		err     error
		reply   []byte
		logList []string
		reqBody JobReqParams
	)

	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Exec", reqBody.JobID, &reply); err != nil {
		goto failed
	}

	ctx.pubEvent(event_ExecCronJob, reqBody.Addr, reqBody)
	logList = strings.Split(string(reply), "\n")
	ctx.respSucc("", logList)
	return

failed:
	ctx.respError(proto.Code_Error, err.Error(), nil)
}

func getJob(c iris.Context) {
	var (
		ctx        = wrapCtx(c)
		reqBody    GetJobReqParams
		crontabJob models.CrontabJob
		err        error
	)
	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Get", reqBody.JobID, &crontabJob); err != nil {
		goto failed
	}

	ctx.respSucc("", crontabJob)
	return
failed:
	ctx.respError(proto.Code_Error, err.Error(), nil)
}
