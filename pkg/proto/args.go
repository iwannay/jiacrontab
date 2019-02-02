package proto

var Months = [...]string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

var Days = [...]string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

type MailArgs struct {
	Host string
	User string
	Pass string
	Port string
}

type SearchLog struct {
	JobID          int
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
