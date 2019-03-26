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

type QueryJobArgs struct{ Page, Pagesize int }

type QueryCrontabJobRet struct {
	Total    int
	Page     int
	Pagesize int
	List     []models.CrontabJob
}

type QueryDaemonJobRet struct {
	Total    int
	Page     int
	Pagesize int
	List     []models.DaemonJob
}

type AuditJobArgs struct{ JobIDs []uint }

type CrontabApiNotifyBody struct {
	NodeAddr  string
	JobName   string
	JobID     int
	Commands  [][]string
	CreatedAt time.Time
	Timeout   int64
	Type      string
	RetryNum  int
}
