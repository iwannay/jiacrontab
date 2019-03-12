package proto

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
	Url  string
	Data string
}

type EmptyArgs struct{}

type EmptyReply struct{}
