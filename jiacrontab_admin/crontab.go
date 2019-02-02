package admin

import (
	"fmt"
	"jiacrontab/model"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
	"net/http"
	"strings"
	"time"

	"github.com/kataras/iris"
)

func getJobList(c iris.Context) {

	var (
		systemInfo map[string]interface{}
		jobList    []models.CrontabJob
		clientList []models.Client
		client     models.Client
		ctx        = wrapCtx(c)
		addr       = ctx.FormValue("addr")
	)

	if strings.TrimSpace(addr) == "" {
		ctx.respError(proto.Code_Error, "参数错误", nil)
		return
	}

	model.DB().Model(&models.Client{}).Find(&clientList)

	if len(clientList) == 0 {
		ctx.respError(proto.Code_Error, "暂无数据", nil)
		return
	}

	for _, v := range clientList {
		if v.Addr == addr {
			client = v
			break
		}
	}

	if err := rpc.Call(addr, "CrontabJob.All", "", &jobList); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	if err := rpc.Call(addr, "Admin.SystemInfo", "", &systemInfo); err != nil {
		ctx.respError(proto.Code_Error, err.Error(), nil)
		return
	}

	ctx.respSucc("", map[string]interface{}{
		"taskList":   jobList,
		"addr":       addr,
		"client":     client,
		"clientList": clientList,
		"systemInfo": systemInfo,
	})
}

func getRecentLog(c iris.Context) {
	var (
		err       error
		ctx       = wrapCtx(c)
		searchRet proto.SearchLogResult
		reqBody   getLogReqParams
	)


	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if err := rpc.Call(reqBody.Addr, "CrontabTask.Log", proto.SearchLog{
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

func editJob(c iris.Context) {
	var (
		err error
		ctx     = wrapCtx(c)
		reqBody editJobReqParams
		rpcArgs models.CrontabJob
	)

	if err = reqBody.verify(ctx);err != nil {
		goto failed
	}
	

	rpcArgs = models.CrontabJob{
		Name:     reqBody.Name,
		Commands: reqBody.Commands,
		TimeArgs: {
			Month:   reqBody.Month,
			Day:     reqBody.Day,
			Hour:    reqBody.Hour,
			Minute:  reqBody.Minute,
			Weekday: reqBody.Weekday,
		},
		PipeCommands:    reqBody.PipeCommands,
		Timeout:         reqBody.ExecuteTimeout,
		TimeoutTrigger:  reqBody.TimeoutTrigger,
		Create:          time.Now().Unix(),
		MailTo:          reqBody.MailTo,
		ApiTo:           reqBody.APITo,
		MaxConcurrent:   reqBody.MaxConcurrent,
		DependJobs:      reqBody.DependJobs,
		ErrorMailNotify: reqBody.ErrorMailNotify,
		ErrorAPINotify:  reqBody.ErrorAPINotify,
		IsSync:          reqBody.IsSync,
	}

	rpcArgs.ID = reqBody.ID

	if err := rpcCall(addr, "CrontabJob.Update", rpcArgs, &reply); err != nil {
		goto failed
	}
	return ctx.respSucc("", nil)

	failed:
	return ctx.respError(proto.Code_Error, err.Error(), nil)
}

func stopTask(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		err     error
		reply   bool
		ok bool
		method string
		reqBody stopTaskReqParams
		methods = map[string]string{
			"stop":"crontabTask.Stop",
			"delete":"CrontabTask.Delete",
			"kill":"CrontabTask.Kill"
		}
	)

	if reqBody.verify(ctx);err != nil {
		goto failed
	}

	if method,ok = methods[reqBody.Action];!ok {
		err = errors.New("action无法识别")
		goto failed
	}


	if err := rpcCall(addr, method, reqBody.JobID, &reply); err != nil {
		goto failed
	}

	return ctx.respSucc("", reply)

	failed:
	return	ctx.respError(proto.Code_Error,err.Error(),nil)

}

func startTask(c iris.Context) {
	var (
		ctx = wrapCtx(c)
		reply bool
		reqBody startTaskReqParams
	)

	if err = reqBody.verify(ctx);err != nil {
		goto failed
	}

	if err := rpcCall(addr, "CrontabTask.Start", taskId, &reply); err != nil {
		
		goto failed
	}

	return ctx.respSucc("", reply)
failed:
	return ctx.respError(proto.Code_Error,err.Error(), nil)
}

func execTask(c iris.Context) {
	var (
		ctx = wrapCtx(c)
		err error
		reply bool
		logList []string
		reqBody execTaskReqParams
	)

	if err = reqBody.verify(ctx);err != nil {
		goto failed
	}

	if err = rpcCall(reqBody.Addr, "CrontabTask.QuickStart",reqBody.JobID, &reply); err != nil {
		goto failed
	}

	logList = strings.Split(string(reply),"\n")
	return ctx.respSucc("", logList)

failed:
	return ctx.respError(proto.Code_Error,err.Error(), nil)
}