package proto

type Client struct {
	Name           string
	State          int
	DaemonTaskNum  int
	CrontabTaskNum int
	Addr           string
	Mail           string
}
