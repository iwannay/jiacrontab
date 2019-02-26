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
		reqBody      getJobListReqParams
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
		reqBody   getLogReqParams
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
		reqBody editJobReqParams
		rpcArgs models.CrontabJob
	)

	if err = reqBody.verify(ctx); err != nil {
		goto failed
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
		// PipeCommands:    reqBody.PipeCommands,
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
		reqBody actionTaskReqParams
		methods = map[string]string{
			"stop":   "CrontabJob.Stop",
			"delete": "CrontabJob.Delete",
			"kill":   "CrontabJob.Kill",
		}
		eDesc = map[string]string{
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

	if err := rpcCall(reqBody.Addr, method, reqBody.JobIDs, &reply); err != nil {
		goto failed
	}

	ctx.pubEvent(eDesc[reqBody.Action], reqBody.Addr, reqBody)
	ctx.respSucc("", reply)
	return

failed:
	ctx.respError(proto.Code_Error, err.Error(), nil)

}

func startTask(c iris.Context) {
	var (
		err     error
		ctx     = wrapCtx(c)
		reply   bool
		reqBody jobsReqParams
	)

	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if err = rpcCall(reqBody.Addr, "CrontabJob.Start", reqBody.JobIDs, &reply); err != nil {
		goto failed
	}

	ctx.pubEvent(event_StartCronJob, reqBody.Addr, reqBody)
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
		reqBody jobReqParams
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
		reqBody    getJobReqParams
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
