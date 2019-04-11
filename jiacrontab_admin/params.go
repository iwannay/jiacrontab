package admin

import (
	"errors"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"

	"github.com/iwannay/log"

	"github.com/kataras/iris"
)

var (
	paramsError = errors.New("参数错误")
)

type Parameter interface {
	Verify(iris.Context) error
}

type JobReqParams struct {
	JobID uint   `json:"jobID"`
	Addr  string `json:"addr"`
}

func (p *JobReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.JobID == 0 || p.Addr == "" {
		return paramsError
	}
	return nil
}

type JobsReqParams struct {
	JobIDs []uint `json:"jobIDs"`
	Addr   string `json:"addr"`
}

func (p *JobsReqParams) verify(ctx iris.Context) error {
	if ctx != nil {
		if err := ctx.ReadJSON(p); err != nil {
			return paramsError
		}
	}

	if len(p.JobIDs) == 0 || p.Addr == "" {
		return paramsError
	}
	return nil
}

type EditJobReqParams struct {
	JobID           uint              `json:"jobID"`
	Addr            string            `json:"addr"`
	IsSync          bool              `json:"isSync"`
	Name            string            `json:"name"`
	Commands        [][]string        `json:"commands"`
	Timeout         int               `json:"timeout"`
	MaxConcurrent   uint              `json:"maxConcurrent"`
	ErrorMailNotify bool              `json:"errorMailNotify"`
	ErrorAPINotify  bool              `json:"errorAPINotify"`
	MailTo          []string          `json:"mailTo"`
	APITo           []string          `json:"APITo"`
	RetryNum        int               `json:"retryNum"`
	WorkDir         string            `json:"workDir"`
	WorkUser        string            `json:"workUser"`
	WorkEnv         []string          `json:"workEnv"`
	DependJobs      models.DependJobs `json:"dependJobs"`
	Month           string            `json:"month"`
	Weekday         string            `json:"weekday"`
	Day             string            `json:"day"`
	Hour            string            `json:"hour"`
	Minute          string            `json:"minute"`
	Second          string            `json:"second"`
	TimeoutTrigger  []string          `json:"timeoutTrigger"`
}

// TODO:验证参数
func (p *EditJobReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return err
	}
	ts := map[string]bool{
		proto.TimeoutTrigger_CallApi:   true,
		proto.TimeoutTrigger_SendEmail: true,
		proto.TimeoutTrigger_Kill:      true,
	}

	for _, v := range p.TimeoutTrigger {
		if !ts[v] {
			return fmt.Errorf("%s:%v", v, paramsError)
		}
	}

	if p.Month == "" {
		p.Month = "*"
	}

	if p.Weekday == "" {
		p.Weekday = "*"
	}

	if p.Day == "" {
		p.Day = "*"
	}

	if p.Hour == "" {
		p.Hour = "*"
	}

	if p.Minute == "" {
		p.Minute = "*"
	}

	if p.Second == "" {
		p.Second = "*"
	}

	return nil
}

type GetLogReqParams struct {
	Addr     string `json:"addr"`
	JobID    uint   `json:"jobID"`
	Date     string `json:"date"`
	Pattern  string `json:"pattern"`
	IsTail   bool   `json:"isTail"`
	Page     int    `json:"page"`
	Pagesize int    `json:"pagesize"`
}

func (p *GetLogReqParams) verify(ctx iris.Context) error {
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

type DeleteNodeReqParams struct {
	Addr    string `json:"addr"`
	GroupID uint   `json:"groupID"`
}

func (p *DeleteNodeReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return paramsError
	}
	return nil
}

type SendTestMailReqParams struct {
	MailTo string `json:"mailTo"`
}

func (p *SendTestMailReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.MailTo == "" {
		return paramsError
	}
	return nil
}

type SystemInfoReqParams struct {
	Addr string `json:"addr"`
}

func (p *SystemInfoReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return paramsError
	}
	return nil
}

type GetJobListReqParams struct {
	Addr string `json:"addr"`
	PageReqParams
}

func (p *GetJobListReqParams) verify(ctx iris.Context) error {
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

type GetGroupListReqParams struct {
	PageReqParams
}

func (p *GetGroupListReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil {
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

type ActionTaskReqParams struct {
	Action string `json:"action"`
	Addr   string `json:"addr"`
	JobIDs []uint `json:"jobIDs"`
}

func (p *ActionTaskReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" ||
		p.Action == "" || len(p.JobIDs) == 0 {
		return paramsError
	}
	return nil
}

type EditDaemonJobReqParams struct {
	Addr            string   `json:"addr"`
	JobID           uint     `json:"jobID"`
	Name            string   `json:"name"`
	MailTo          []string `json:"mailTo"`
	APITo           []string `json:"APITo"`
	Commands        []string `json:"commands"`
	WorkUser        string   `json:"workUser"`
	WorkEnv         []string `json:"workEnv"`
	WorkDir         string   `json:"workDir"`
	FailRestart     bool     `json:"failRestart"`
	ErrorMailNotify bool     `json:"errorMailNotify"`
	ErrorAPINotify  bool     `json:"errorAPINotify"`
}

func (p *EditDaemonJobReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" || p.Name == "" ||
		len(p.Commands) == 0 {
		return paramsError
	}
	return nil
}

type GetJobReqParams struct {
	JobID uint   `json:"jobID"`
	Addr  string `json:"addr"`
}

func (p *GetJobReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.JobID == 0 || p.Addr == "" {
		return paramsError
	}
	return nil
}

type UserReqParams struct {
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
	GroupID  uint   `json:"groupID"`
	Root     bool   `json:"root"`
	Mail     string `json:"mail"`
}

func (p *UserReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Username == "" || p.Passwd == "" {
		return paramsError
	}

	return nil
}

type LoginReqParams struct {
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
	Remember bool   `json:"remember"`
}

func (p *LoginReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Username == "" || p.Passwd == "" {
		return paramsError
	}

	return nil
}

type PageReqParams struct {
	Page     int `json:"page"`
	Pagesize int `json:"pagesize"`
}

type GetNodeListReqParams struct {
	PageReqParams
}

func (p *GetNodeListReqParams) verify(ctx iris.Context) error {
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

type EditGroupReqParams struct {
	GroupID   uint   `json:"groupID"`
	GroupName string `json:"groupName"`
}

func (p *EditGroupReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.GroupName == "" {
		return paramsError
	}
	return nil
}

type SetGroupReqParams struct {
	TargetGroupID uint `json:"targetGroupID"`
	UserID        uint `json:"userID"`
	Root          bool `json:"root"`
}

func (p *SetGroupReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.UserID == 0 {
		return paramsError
	}
	return nil
}

type ReadMoreReqParams struct {
	LastID   int    `json:"lastID"`
	Pagesize uint   `json:"pagesize"`
	Orderby  string `json:"orderby"`
}

func (p *ReadMoreReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil {
		return paramsError
	}

	if p.Pagesize == 0 {
		p.Pagesize = 50
	}

	if p.Orderby == "" {
		p.Orderby = "desc"
	}

	return nil
}

type GroupNodeReqParams struct {
	Addr            string `json:"addr"`
	TargetNodeName  string `json:"targetNodeName"`
	TargetGroupName string `json:"targetGroupName"`
	TargetGroupID   uint   `json:"targetGroupID"`
}

func (p *GroupNodeReqParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" ||
		(p.TargetGroupID == 0 && p.TargetGroupName == "") {
		return paramsError
	}
	return nil
}

type AuditJobReqParams struct {
	JobsReqParams
	JobType string `json:"jobType"`
}

func (p *AuditJobReqParams) verify(ctx iris.Context) error {

	if err := ctx.ReadJSON(p); err != nil || p.Addr == "" {
		return paramsError
	}

	log.Debugf("%+v", p)

	jobTypeMap := map[string]bool{
		"crontab": true,
		"daemon":  true,
	}

	if err := p.JobsReqParams.verify(nil); err != nil {
		return err
	}

	if jobTypeMap[p.JobType] == false {
		return paramsError
	}

	return nil
}

type GetUsersParams struct {
	PageReqParams
	QueryGroupID uint `json:"queryGroupID"`
}

func (p *GetUsersParams) verify(ctx iris.Context) error {
	if err := ctx.ReadJSON(p); err != nil {
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
