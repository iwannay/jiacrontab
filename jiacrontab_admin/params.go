package admin

import (
	"fmt"
	"jiacrontab/models"
)

type execTaskReqParams struct {
	JobID int    `json:"jobID"`
	Addr  string `json:"addr"`
}

func (p *execTaskReqParams) verify(ctx gin.Context) error {
	if err = ctx.ReadJSON(p); err != nil || p.JobID == 0 || p.Addr == "" {
		return fmt.Errorf("参数错误")
	}

	return nil
}

type startTaskReqParams struct {
	JobID int    `json:"jobID"`
	Addr  string `json:"addr"`
}

func (p *startTaskReqParams) verify(ctx gin.Context) error {
	if err = ctx.ReadJSON(p); err != nil || p.JobID == 0 || p.Addr == "" {
		return fmt.Errorf("参数错误")
	}

	return nil
}

type stopTaskReqParams struct {
	JobID  int    `json:"jobID"`
	Addr   string `json:"addr"`
	Action string `json:"action"`
}

func (p *stopTaskReqParams) verify(ctx gin.Context) error {
	if err = ctx.ReadJSON(p); err != nil ||
		p.JobID == 0 || p.Addr == "" || p.Action {
		return fmt.Errorf("参数错误")
	}

	return nil
}

type editJobReqParams struct {
	ID              int               `json:"id"`
	Addr            string            `json:"addr"`
	IsSync          bool              `json:"isSync"`
	Commands        []string          `json:"command"`
	Args            string            `json:"args"`
	Name            string            `json:"name"`
	Timeout         string            `json:"timeout"`
	MaxConcurrent   string            `json:"maxConcurrent"`
	ErrorMailNotify bool              `json:"mailNotify"`
	ErrorAPINotify  bool              `json:"APINotify"`
	MailTo          string            `json:"mailTo"`
	APITo           string            `json:"APITo"`
	Timeout         string            `json:"timeout"`
	PipeCommands    [][]string        `json:"pipeCommands"`
	DependJobs      models.DependJobs `json:"dependents"`
	Month           string            `json:"month"`
	Weekday         string            `json:"weekday"`
	Day             string            `json:"day"`
	Hour            string            `json:"hour"`
	Minute          string            `json:"minute"`
	TimeoutTrigger  string            `json:"timeoutTrigger"`
}

func (p *editJobReqParams) verify(ctx gin.Context) error {
	if err = ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return fmt.Errorf("参数错误")
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

func (p *getLogReqParams) verify(ctx gin.Context) error {
	if err = ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return fmt.Errorf("参数错误")
	}

	if p.Page == 0 {
		p.Page = 1
	} else if p.Pagesize <= 0 {
		p.Pagesize = 50
	}

	return nil
}
