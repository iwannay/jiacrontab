package proto

import (
	"jiacrontab/models"
)

type SearchLog struct {
	JobID          uint
	IsTail         bool
	Page, Pagesize int
	Date           string
	Pattern        string
}

type SearchLogResult struct {
	Content []byte
	Total   int
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
