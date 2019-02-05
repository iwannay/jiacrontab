package admin

import (
	"errors"
	"jiacrontab/models"

	"github.com/kataras/iris"
)

var (
	paramsError = errors.New("参数错误")
)

type execTaskReqParams struct {
	JobID uint   `json:"jobID"`
	Addr  string `json:"addr"`
}

func (p *execTaskReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.JobID == 0 || p.Addr == "" {
		return paramsError
	}

	return nil
}

type startTaskReqParams struct {
	JobID uint   `json:"jobID"`
	Addr  string `json:"addr"`
}

func (p *startTaskReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.JobID == 0 || p.Addr == "" {
		return paramsError
	}

	return nil
}

type stopTaskReqParams struct {
	JobIDs []int  `json:"jobID"`
	Addr   string `json:"addr"`
	Action string `json:"action"`
}

func (p *stopTaskReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil ||
		len(p.JobIDs) == 0 || p.Addr == "" || p.Action == "" {
		return paramsError
	}

	return nil
}

type editJobReqParams struct {
	ID              uint              `json:"id"`
	Addr            string            `json:"addr"`
	IsSync          bool              `json:"isSync"`
	Commands        []string          `json:"command"`
	Args            string            `json:"args"`
	Name            string            `json:"name"`
	Timeout         int               `json:"timeout"`
	MaxConcurrent   uint              `json:"maxConcurrent"`
	ErrorMailNotify bool              `json:"mailNotify"`
	ErrorAPINotify  bool              `json:"APINotify"`
	MailTo          string            `json:"mailTo"`
	APITo           string            `json:"APITo"`
	PipeCommands    [][]string        `json:"pipeCommands"`
	DependJobs      models.DependJobs `json:"dependents"`
	Month           string            `json:"month"`
	Weekday         string            `json:"weekday"`
	Day             string            `json:"day"`
	Hour            string            `json:"hour"`
	Minute          string            `json:"minute"`
	TimeoutTrigger  string            `json:"timeoutTrigger"`
}

func (p *editJobReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return paramsError
	}

	return nil
}

type getLogReqParams struct {
	Addr     string `json:"addr"`
	JobID    int    `json:"jobID"`
	Date     string `json:"date"`
	Pattern  string `json:"pattern"`
	IsTail   bool   `json:"isTail"`
	Page     int    `json:"page"`
	Pagesize int    `json:"pagesize"`
}

func (p *getLogReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return paramsError
	}

	if p.Page == 0 {
		p.Page = 1
	} else if p.Pagesize <= 0 {
		p.Pagesize = 50
	}

	return nil
}

type deleteNodeReqParams struct {
	NodeID int    `json:"nodeID"`
	Addr   string `json:"addr"`
}

func (p *deleteNodeReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.NodeID == 0 || p.Addr == "" {
		return paramsError
	}
	return nil
}

type sendTestMailReqParams struct {
	MailTo string `json:"mailTo"`
}

func (p *sendTestMailReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.MailTo == "" {
		return paramsError
	}
	return nil
}

type runtimeInfoReqParams struct {
	Addr string `json:"addr"`
}

func (p *runtimeInfoReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return paramsError
	}
	return nil
}

type jobListReqParams struct {
	Addr     string `json:"addr"`
	Page     int    `json:"page"`
	Pagesize int    `json:"pagesize"`
}

func (p *jobListReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return paramsError
	}

	if p.Page <= 1 {
		p.Page = 1
	}

	if p.Pagesize <= 0 {
		p.Pagesize = 50
	}
	return nil
}

type actionTaskReqParams struct {
	Action string `json:"action"`
	Addr   string `json:"addr"`
	JobIDs []uint `json:"jobIDs"`
}

func (p *actionTaskReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" ||
		p.Action == "" || len(p.JobIDs) == 0 {
		return paramsError
	}
	return nil
}

type editDaemonJobReqParams struct {
	Addr            string   `json:"addr"`
	JobID           int      `json:"jobID"`
	Name            string   `json:"name"`
	MailTo          string   `json:"mailTo"`
	APITo           string   `json:"apiTo"`
	Commands        []string `json:"commands"`
	FailRestart     bool     `json:"failRestart"`
	ErrorMailNotify bool     `json:"mailNotify"`
	ErrorAPINotify  bool     `json:"APINotify"`
}

func (p *editDaemonJobReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" || p.Name == "" ||
		len(p.Commands) == 0 {
		return paramsError
	}
	return nil
}

type getJobReqParams struct {
	JobID uint   `json:"jobID"`
	Addr  string `json:"addr"`
}

func (p *getJobReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.JobID == 0 || p.Addr == "" {
		return paramsError
	}
	return nil
}

type userReqParams struct {
	Name    string `json:"name"`
	Passwd  string `json:"string"`
	GroupID int    `json:"group"`
	Root    bool   `json:"root"`
	Email   string `json:"email"`
}

func (p *userReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Name == "" || p.Passwd == "" {
		return paramsError
	}

	return nil
}
