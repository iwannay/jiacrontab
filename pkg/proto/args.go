package proto

import (
	"jiacrontab/models"
)

type SearchLog struct {
	JobID    uint
	GroupID  uint
	UserID   uint
	Root     bool
	IsTail   bool
	Offset   int64
	Pagesize int
	Date     string
	Pattern  string
}

type CleanNodeLog struct {
	Unit   string
	Offset int
}

type CleanNodeLogRet struct {
	Total int64  `json:"total"`
	Size  string `json:"size"`
}

type SearchLogResult struct {
	Content  []byte
	Offset   int64
	FileSize int64
}

type SendMail struct {
	MailTo  []string
	Subject string
	Content string
}

type ApiPost struct {
	Urls []string
	Data string
}

type ExecCrontabJobReply struct {
	Job     models.CrontabJob
	Content []byte
}

type ActionJobsArgs struct {
	UserID  uint
	Root    bool
	GroupID uint
	JobIDs  []uint
}

type GetJobArgs struct {
	UserID  uint
	GroupID uint
	Root    bool
	JobID   uint
}
type EmptyArgs struct{}

type EmptyReply struct{}
