package proto

type TaskArgs struct {
	Id            string
	Name          string
	Command       string
	State         int // 0/1/2
	Args          string
	Create        int64
	LastCostTime  int64
	LastExecTime  int64
	Timeout       int64
	NumberProcess int32
	MailTo        string
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
