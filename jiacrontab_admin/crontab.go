package admin

import (
	"errors"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"strings"

	"github.com/kataras/iris"
)

func GetJobList(c iris.Context) {

	var (
		jobRet       proto.QueryCrontabJobRet
		ctx          = wrapCtx(c)
		err          error
		reqBody      GetJobListReqParams
		rpcReqParams proto.QueryJobArgs
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	rpcReqParams.Page = reqBody.Page
	rpcReqParams.Pagesize = reqBody.Pagesize

	if err := rpc.Call(reqBody.Addr, "CrontabJob.List", rpcReqParams, &jobRet); err != nil {
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

func GetRecentLog(c iris.Context) {
	var (
		err       error
		ctx       = wrapCtx(c)
		searchRet proto.SearchLogResult
		reqBody   GetLogReqParams
		logList   []string
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if err = rpc.Call(reqBody.Addr, "CrontabJob.Log", proto.SearchLog{
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
		"fileSize": searchRet.FileSize,
		"pagesize": reqBody.Pagesize,
	})
}

func EditJob(c iris.Context) {
	var (
		err     error
		reply   models.CrontabJob
		ctx     = wrapCtx(c)
		reqBody EditJobReqParams
		rpcArgs models.CrontabJob
	)

	if err = ctx.parseClaimsFromToken(); err != nil {
		ctx.respJWTError(err)
		return
	}

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respBasicError(err)
		return
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
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

		UpdatedUserID:   ctx.claims.UserID,
		UpdatedUsername: ctx.claims.Username,
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

	rpcArgs.ID = reqBody.JobID

	if rpcArgs.ID == 0 {
		rpcArgs.CreatedUserID = ctx.claims.UserID
		rpcArgs.CreatedUsername = ctx.claims.Username
	}

	if ctx.claims.Root || ctx.claims.GroupID == models.SuperGroup.ID {
		rpcArgs.Status = models.StatusJobOk
	} else {
		rpcArgs.Status = models.StatusJobUnaudited
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Edit", rpcArgs, &reply); err != nil {
		ctx.respRPCError(err)
		return
	}
	ctx.pubEvent(reply.Name, event_EditCronJob, reqBody.Addr, reqBody)
	ctx.respSucc("", reply)
}

func ActionTask(c iris.Context) {
	var (
		ctx      = wrapCtx(c)
		err      error
		reply    bool
		ok       bool
		method   string
		reqBody  ActionTaskReqParams
		jobReply []models.CrontabJob
		methods  = map[string]string{
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

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respBasicError(err)
		return
	}

	if method, ok = methods[reqBody.Action]; !ok {
		ctx.respBasicError(errors.New("action无法识别"))
		return
	}

	if err = rpcCall(reqBody.Addr, method, reqBody.JobIDs, &jobReply); err != nil {
		ctx.respRPCError(err)
		return
	}
	var targetNames []string
	for _, v := range jobReply {
		targetNames = append(targetNames, v.Name)
	}
	ctx.pubEvent(strings.Join(targetNames, ","), eDesc[reqBody.Action], reqBody.Addr, reqBody)
	ctx.respSucc("", reply)
}

func ExecTask(c iris.Context) {
	var (
		ctx          = wrapCtx(c)
		err          error
		logList      []string
		execJobReply proto.ExecCrontabJobReply
		reqBody      JobReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Exec", reqBody.JobID, &execJobReply); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.pubEvent(execJobReply.Job.Name, event_ExecCronJob, reqBody.Addr, reqBody)
	logList = strings.Split(string(execJobReply.Content), "\n")
	ctx.respSucc("", logList)
}

func GetJob(c iris.Context) {
	var (
		ctx        = wrapCtx(c)
		reqBody    GetJobReqParams
		crontabJob models.CrontabJob
		err        error
	)

	if err = ctx.parseClaimsFromToken(); err != nil {
		ctx.respJWTError(err)
		return
	}

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Get", reqBody.JobID, &crontabJob); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.respSucc("", crontabJob)
}
