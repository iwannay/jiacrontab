package admin

import (
	"errors"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/proto"

	"github.com/kataras/iris"
)

var (
	paramsError = errors.New("参数错误")
)

type Parameter interface {
	Verify(iris.Context) error
}

type JobReqParams struct {
	JobID uint   `json:"jobID" rule:"required,请填写jobID"`
	Addr  string `json:"addr"  rule:"required,请填写addr"`
}

func (p *JobReqParams) Verify(ctx iris.Context) error {
	if p.JobID == 0 || p.Addr == "" {
		return paramsError
	}
	return nil
}

type JobsReqParams struct {
	JobIDs []uint `json:"jobIDs" `
	Addr   string `json:"addr"`
}

func (p *JobsReqParams) Verify(ctx iris.Context) error {

	if len(p.JobIDs) == 0 || p.Addr == "" {
		return paramsError
	}

	return nil
}

type EditJobReqParams struct {
	JobID            uint              `json:"jobID"`
	Addr             string            `json:"addr" rule:"required,请填写addr"`
	IsSync           bool              `json:"isSync"`
	Name             string            `json:"name" rule:"required,请填写name"`
	Command          []string          `json:"command" rule:"required,请填写name"`
	Code             string            `json:"code"`
	Timeout          int               `json:"timeout"`
	MaxConcurrent    uint              `json:"maxConcurrent"`
	ErrorMailNotify  bool              `json:"errorMailNotify"`
	ErrorAPINotify   bool              `json:"errorAPINotify"`
	MailTo           []string          `json:"mailTo"`
	APITo            []string          `json:"APITo"`
	RetryNum         int               `json:"retryNum"`
	WorkDir          string            `json:"workDir"`
	WorkUser         string            `json:"workUser"`
	WorkEnv          []string          `json:"workEnv"`
	KillChildProcess bool              `json:"killChildProcess"`
	DependJobs       models.DependJobs `json:"dependJobs"`
	Month            string            `json:"month"`
	Weekday          string            `json:"weekday"`
	Day              string            `json:"day"`
	Hour             string            `json:"hour"`
	Minute           string            `json:"minute"`
	Second           string            `json:"second"`
	TimeoutTrigger   []string          `json:"timeoutTrigger"`
}

func (p *EditJobReqParams) Verify(ctx iris.Context) error {
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
	Offset   int64  `json:"offset"`
	Pagesize int    `json:"pagesize"`
}

func (p *GetLogReqParams) Verify(ctx iris.Context) error {

	if p.Pagesize <= 0 {
		p.Pagesize = 50
	}

	return nil
}

type DeleteNodeReqParams struct {
	Addr    string `json:"addr" rule:"required,请填写addr"`
	GroupID uint   `json:"groupID"`
}

func (p *DeleteNodeReqParams) Verify(ctx iris.Context) error {
	return nil
}

type SendTestMailReqParams struct {
	MailTo string `json:"mailTo" rule:"required,请填写mailTo"`
}

func (p *SendTestMailReqParams) Verify(ctx iris.Context) error {
	return nil
}

type SystemInfoReqParams struct {
	Addr string `json:"addr" rule:"required,请填写addr"`
}

func (p *SystemInfoReqParams) Verify(ctx iris.Context) error {
	return nil
}

type GetJobListReqParams struct {
	Addr      string `json:"addr" rule:"required,请填写addr"`
	SearchTxt string `json:"searchTxt"`
	PageReqParams
}

func (p *GetJobListReqParams) Verify(ctx iris.Context) error {

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

func (p *GetGroupListReqParams) Verify(ctx iris.Context) error {

	if p.Page <= 1 {
		p.Page = 1
	}

	if p.Pagesize <= 0 {
		p.Pagesize = 50
	}
	return nil
}

type ActionTaskReqParams struct {
	Action string `json:"action" rule:"required,请填写action"`
	Addr   string `json:"addr" rule:"required,请填写addr"`
	JobIDs []uint `json:"jobIDs" rule:"required,请填写jobIDs"`
}

func (p *ActionTaskReqParams) Verify(ctx iris.Context) error {
	if len(p.JobIDs) == 0 {
		return paramsError
	}
	return nil
}

type EditDaemonJobReqParams struct {
	Addr            string   `json:"addr" rule:"required,请填写addr"`
	JobID           uint     `json:"jobID"`
	Name            string   `json:"name" rule:"required,请填写name"`
	MailTo          []string `json:"mailTo"`
	APITo           []string `json:"APITo"`
	Command         []string `json:"command"  rule:"required,请填写command"`
	Code            string   `json:"code"`
	WorkUser        string   `json:"workUser"`
	WorkEnv         []string `json:"workEnv"`
	WorkDir         string   `json:"workDir"`
	FailRestart     bool     `json:"failRestart"`
	RetryNum        int      `json:"retryNum"`
	ErrorMailNotify bool     `json:"errorMailNotify"`
	ErrorAPINotify  bool     `json:"errorAPINotify"`
}

func (p *EditDaemonJobReqParams) Verify(ctx iris.Context) error {
	return nil
}

type GetJobReqParams struct {
	JobID uint   `json:"jobID" rule:"required,请填写jobID"`
	Addr  string `json:"addr" rule:"required,请填写addr"`
}

func (p *GetJobReqParams) Verify(ctx iris.Context) error {
	return nil
}

type UserReqParams struct {
	Username  string `json:"username" rule:"required,请输入用户名"`
	Passwd    string `json:"passwd,omitempty" rule:"required,请输入密码"`
	GroupID   uint   `json:"groupID"`
	GroupName string `json:"groupName"`
	Avatar    string `json:"avatar"`
	Root      bool   `json:"root"`
	Mail      string `json:"mail"`
}

func (p *UserReqParams) Verify(ctx iris.Context) error {
	return nil
}

type InitAppReqParams struct {
	Username string `json:"username" rule:"required,请输入用户名"`
	Passwd   string `json:"passwd" rule:"required,请输入密码"`
	Avatar   string `json:"avatar"`
	Mail     string `json:"mail"`
}

func (p *InitAppReqParams) Verify(ctx iris.Context) error {
	return nil
}

type EditUserReqParams struct {
	UserID   uint   `json:"userID" rule:"required,缺少userID"`
	Username string `json:"username" rule:"required,请输入用户名"`
	Passwd   string `json:"passwd" rule:"required,请输入密码"`
	GroupID  uint   `json:"groupID" rule:"required,请输入密码"`
	Avatar   string `json:"avatar"`
	Mail     string `json:"mail"`
}

func (p *EditUserReqParams) Verify(ctx iris.Context) error {
	return nil
}

type LoginReqParams struct {
	Username string `json:"username" rule:"required,请输入用户名"`
	Passwd   string `json:"passwd" rule:"required,请输入密码"`
	Remember bool   `json:"remember"`
}

func (p *LoginReqParams) Verify(ctx iris.Context) error {
	return nil
}

type PageReqParams struct {
	Page     int `json:"page"`
	Pagesize int `json:"pagesize"`
}

type GetNodeListReqParams struct {
	PageReqParams
	QueryGroupID uint `json:"queryGroupID"`
}

func (p *GetNodeListReqParams) Verify(ctx iris.Context) error {

	if p.Page == 0 {
		p.Page = 1
	}
	if p.Pagesize <= 0 {
		p.Pagesize = 50
	}
	return nil
}

type EditGroupReqParams struct {
	GroupID   uint   `json:"groupID" rule:"required,请填写groupID"`
	GroupName string `json:"groupName"  rule:"required,请填写groupName"`
}

func (p *EditGroupReqParams) Verify(ctx iris.Context) error {
	return nil
}

type SetGroupReqParams struct {
	TargetGroupID   uint   `json:"targetGroupID"`
	TargetGroupName string `json:"targetGroupName"`
	UserID          uint   `json:"userID" rule:"required,请填写用户ID"`
	Root            bool   `json:"root"`
}

func (p *SetGroupReqParams) Verify(ctx iris.Context) error {
	return nil
}

type ReadMoreReqParams struct {
	LastID   int    `json:"lastID"`
	Pagesize uint   `json:"pagesize"`
	Orderby  string `json:"orderby"`
}

func (p *ReadMoreReqParams) Verify(ctx iris.Context) error {
	if p.Pagesize == 0 {
		p.Pagesize = 50
	}

	if p.Orderby == "" {
		p.Orderby = "desc"
	}

	return nil
}

type GroupNodeReqParams struct {
	Addr            string `json:"addr" rule:"required,请填写addr"`
	TargetNodeName  string `json:"targetNodeName"`
	TargetGroupName string `json:"targetGroupName"`
	TargetGroupID   uint   `json:"targetGroupID"`
}

func (p *GroupNodeReqParams) Verify(ctx iris.Context) error {
	return nil
}

type AuditJobReqParams struct {
	JobsReqParams
	JobType string `json:"jobType"`
}

func (p *AuditJobReqParams) Verify(ctx iris.Context) error {

	if p.Addr == "" {
		return paramsError
	}

	jobTypeMap := map[string]bool{
		"crontab": true,
		"daemon":  true,
	}

	if err := p.JobsReqParams.Verify(nil); err != nil {
		return err
	}

	if jobTypeMap[p.JobType] == false {
		return paramsError
	}

	return nil
}

type GetUsersParams struct {
	PageReqParams
	IsAll        bool `json:"isAll"`
	QueryGroupID uint `json:"queryGroupID"`
}

func (p *GetUsersParams) Verify(ctx iris.Context) error {

	if p.Page <= 1 {
		p.Page = 1
	}

	if p.Pagesize <= 0 {
		p.Pagesize = 50
	}
	return nil
}
