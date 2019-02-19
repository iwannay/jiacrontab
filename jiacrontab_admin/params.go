package admin

import (
	"errors"
	"jiacrontab/models"

	"github.com/kataras/iris"
)

var (
	paramsError = errors.New("参数错误")
)

type jobReqParams struct {
	JobID uint   `json:"jobID"`
	Addr  string `json:"addr"`
}

func (p *jobReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.JobID == 0 || p.Addr == "" {
		return paramsError
	}
	return nil
}

type editJobReqParams struct {
	ID              uint       `json:"id"`
	Addr            string     `json:"addr"`
	IsSync          bool       `json:"isSync"`
	Name            string     `json:"name"`
	Commands        [][]string `json:"commands"`
	Timeout         int        `json:"timeout"`
	MaxConcurrent   uint       `json:"maxConcurrent"`
	ErrorMailNotify bool       `json:"ErrormailNotify"`
	ErrorAPINotify  bool       `json:"ErrorAPINotify"`
	MailTo          []string   `json:"mailTo"`
	APITo           []string   `json:"APITo"`
	// PipeCommands    [][]string        `json:"pipeCommands"`
	DependJobs     models.DependJobs `json:"dependJobs"`
	Month          string            `json:"month"`
	Weekday        string            `json:"weekday"`
	Day            string            `json:"day"`
	Hour           string            `json:"hour"`
	Minute         string            `json:"minute"`
	TimeoutTrigger string            `json:"timeoutTrigger"`
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
	}
	if p.Pagesize <= 0 {
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
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
	GroupID  uint   `json:"groupID"`
	Root     bool   `json:"root"`
	Mail     string `json:"mail"`
}

func (p *userReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Username == "" || p.Passwd == "" {
		return paramsError
	}

	return nil
}

type loginReqParams struct {
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
	Remember bool   `json:"remember"`
}

func (p *loginReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Username == "" || p.Passwd == "" {
		return paramsError
	}

	return nil
}

type pageReqParams struct {
	Page     int `json:"page"`
	Pagesize int `json:"pagesize"`
	GroupID  int `json:"groupID"`
}

func (p *pageReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil {
		return paramsError
	}

	if p.Page == 0 {
		p.Page = 1
	}
	if p.Pagesize <= 0 {
		p.Pagesize = 50
	}
	return nil
}

type editGroupReqParams struct {
	GroupID  uint   `json:"groupID"`
	Name     string `json:"name"`
	NodeAddr string `json:"nodeAddr"`
}

func (p *editGroupReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Name == "" || p.NodeAddr == "" {
		return paramsError
	}
	return nil
}

type setGroupReqParams struct {
	TargetGroupID uint   `json:"targetGroupID"`
	UserID        uint   `json:"userID"`
	NodeAddr      string `json:"nodeAddr"`
}

func (p *setGroupReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || (p.UserID == 0 && p.NodeAddr == "") {
		return paramsError
	}
	return nil
}

type readMoreReqParams struct {
	LastID   uint   `json:"lastID"`
	Pagesize int    `json:"pagesize"`
	Orderby  string `json:"orderby"`
}

func (p *readMoreReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil {
		return paramsError
	}

	if p.Pagesize == 0 {
		p.Pagesize = 20
	}

	if p.Orderby == "" {
		p.Orderby = "desc"
	}

	return nil
}

type updateNodeReqParams struct {
	NodeID uint   `json:"nodeID"`
	Addr   string `json:"addr"`
	Name   string `json:"name"`
}

func (p *updateNodeReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.NodeID == 0 || p.Addr == "" {
		return paramsError
	}
	return nil
}

type auditJobReqParams struct {
	jobReqParams
	JobType string `json:"jobType"`
}

func (p *auditJobReqParams) verify(ctx iris.Context) error {

	jobTypeMap := map[string]bool{
		"crontab": true,
		"daemon":  true,
	}

	if err := p.jobReqParams.verify(ctx); err != nil {
		return err
	}

	if jobTypeMap[p.JobType] == false {
		return paramsError
	}

	return nil
}
