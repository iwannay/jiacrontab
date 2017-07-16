package proto

type TaskArgs struct {
	Id            string
	Name          string
	Command       string
	Depends       []MScript
	State         int // 0/1/2
	Args          string
	Create        int64
	LastCostTime  int64
	LastExecTime  int64
	Timeout       int64
	NumberProcess int32
	MailTo        string
	MaxConcurrent int    // 脚本最大并发量
	OpTimeout     string // email/kill/email_and_kill/ignore
	C             CrontabArgs
}

type CrontabArgs struct {
	Weekday string
	Month   string
	Day     string
	Hour    string
	Minute  string
}

type MScript struct {
	Dest       string
	From       string
	TaskId     string
	Command    string
	Args       string
	Done       bool
	LogContent []byte `json:"-"`
}

type DependsArgs struct {
	TaskId string
	Dpds   []MScript
}

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

type ClientConf struct {
	State int
	Addr  string
	Mail  string
}
type Mdata map[string]*TaskArgs

var Data = make(Mdata)
