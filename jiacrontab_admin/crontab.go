package admin

import (
	"errors"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"strings"
)

func GetJobList(ctx *myctx) {

	var (
		jobRet       proto.QueryCrontabJobRet
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
	rpcReqParams.UserID = ctx.claims.UserID
	rpcReqParams.Root = ctx.claims.Root
	rpcReqParams.SearchTxt = reqBody.SearchTxt

	if err := rpcCall(reqBody.Addr, "CrontabJob.List", rpcReqParams, &jobRet); err != nil {
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

func GetRecentLog(ctx *myctx) {
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

	if err = rpcCall(reqBody.Addr, "CrontabJob.Log", proto.SearchLog{
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

func EditJob(ctx *myctx) {
	var (
		err     error
		reply   models.CrontabJob
		reqBody EditJobReqParams
		job     models.CrontabJob
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respBasicError(err)
		return
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
	}

	job = models.CrontabJob{
		Name:    reqBody.Name,
		Command: reqBody.Command,
		Code:    reqBody.Code,
		TimeArgs: models.TimeArgs{
			Month:   reqBody.Month,
			Day:     reqBody.Day,
			Hour:    reqBody.Hour,
			Minute:  reqBody.Minute,
			Weekday: reqBody.Weekday,
			Second:  reqBody.Second,
		},

		UpdatedUserID:    ctx.claims.UserID,
		UpdatedUsername:  ctx.claims.Username,
		WorkDir:          reqBody.WorkDir,
		WorkUser:         reqBody.WorkUser,
		WorkEnv:          reqBody.WorkEnv,
		KillChildProcess: reqBody.KillChildProcess,
		RetryNum:         reqBody.RetryNum,
		Timeout:          reqBody.Timeout,
		TimeoutTrigger:   reqBody.TimeoutTrigger,
		MailTo:           reqBody.MailTo,
		APITo:            reqBody.APITo,
		MaxConcurrent:    reqBody.MaxConcurrent,
		DependJobs:       reqBody.DependJobs,
		ErrorMailNotify:  reqBody.ErrorMailNotify,
		ErrorAPINotify:   reqBody.ErrorAPINotify,
		IsSync:           reqBody.IsSync,
	}

	job.ID = reqBody.JobID

	if job.ID == 0 {
		job.CreatedUserID = ctx.claims.UserID
		job.CreatedUsername = ctx.claims.Username
	}

	if ctx.claims.Root || ctx.claims.GroupID == models.SuperGroup.ID {
		job.Status = models.StatusJobOk
	} else {
		job.Status = models.StatusJobUnaudited
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Edit", proto.EditCrontabJobArgs{
		Job:  job,
		Root: ctx.claims.Root,
	}, &reply); err != nil {
		ctx.respRPCError(err)
		return
	}
	ctx.pubEvent(reply.Name, event_EditCronJob, models.EventSourceName(reqBody.Addr), reqBody)
	ctx.respSucc("", reply)
}

func ActionTask(ctx *myctx) {
	var (
		err      error
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

	if err = rpcCall(reqBody.Addr, method, proto.ActionJobsArgs{
		UserID: ctx.claims.UserID,
		Root:   ctx.claims.Root,
		JobIDs: reqBody.JobIDs,
	}, &jobReply); err != nil {
		ctx.respRPCError(err)
		return
	}
	if len(jobReply) > 0 {
		var targetNames []string
		for _, v := range jobReply {
			targetNames = append(targetNames, v.Name)
		}
		ctx.pubEvent(strings.Join(targetNames, ","), eDesc[reqBody.Action], models.EventSourceName(reqBody.Addr), reqBody)
	}
	ctx.respSucc("", jobReply)
}

func ExecTask(ctx *myctx) {
	var (
		err          error
		logList      []string
		execJobReply proto.ExecCrontabJobReply
		reqBody      JobReqParams
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Exec", proto.GetJobArgs{
		UserID: ctx.claims.UserID,
		Root:   ctx.claims.Root,
		JobID:  reqBody.JobID,
	}, &execJobReply); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.pubEvent(execJobReply.Job.Name, event_ExecCronJob, models.EventSourceName(reqBody.Addr), reqBody)
	logList = strings.Split(string(execJobReply.Content), "\n")
	ctx.respSucc("", logList)
}

func GetJob(ctx *myctx) {
	var (
		reqBody    GetJobReqParams
		crontabJob models.CrontabJob
		err        error
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if !ctx.verifyNodePermission(reqBody.Addr) {
		ctx.respNotAllowed()
		return
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Get", proto.GetJobArgs{
		UserID: ctx.claims.UserID,
		Root:   ctx.claims.Root,
		JobID:  reqBody.JobID,
	}, &crontabJob); err != nil {
		ctx.respRPCError(err)
		return
	}

	ctx.respSucc("", crontabJob)
}
