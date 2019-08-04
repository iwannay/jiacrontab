package proto

import (
	"jiacrontab/models"
	"time"
)

type DepJobs []DepJob
type DepJob struct {
	Name        string
	Dest        string
	From        string
	ProcessID   int    // 当前主任务进程id
	ID          string // 依赖任务id
	JobID       uint   // 主任务id
	JobUniqueID string // 主任务唯一标志
	Commands    []string
	Timeout     int64
	Err         error
	LogContent  []byte
}

type QueryJobArgs struct {
	SearchTxt      string
	Root           bool
	GroupID        uint
	UserID         uint
	Page, Pagesize int
}

type QueryCrontabJobRet struct {
	Total    int
	Page     int
	GroupID  uint
	Pagesize int
	List     []models.CrontabJob
}

type QueryDaemonJobRet struct {
	Total    int
	GroupID  int
	Page     int
	Pagesize int
	List     []models.DaemonJob
}

type AuditJobArgs struct {
	GroupID uint
	Root    bool
	UserID  uint
	JobIDs  []uint
}

type CrontabApiNotifyBody struct {
	NodeAddr       string
	JobName        string
	JobID          int
	CreateUsername string
	CreatedAt      time.Time
	Timeout        int64
	Type           string
	RetryNum       int
}

type EditCrontabJobArgs struct {
	Job     models.CrontabJob
	UserID  uint
	GroupID uint
	Root    bool
}
