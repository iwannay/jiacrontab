package proto

import (
	"jiacrontab/models"
)

type SearchLog struct {
	JobID    uint
	IsTail   bool
	Offset   int64
	Pagesize int
	Date     string
	Pattern  string
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

type EmptyArgs struct{}

type EmptyReply struct{}
