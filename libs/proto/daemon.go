package proto

const (
	StopDaemonTask = iota
	StartDaemonTask
	DeleteDaemonTask
)

//
//type DaemonTask struct {
//	gorm.Model
//	Name       string `gorm:"unique;not null"`
//	MailNofity bool
//	Status     int
//	MailTo     string
//	Command    string
//	Args       string
//}

type ActionDaemonTaskArgs struct {
	Action int
	TaskId int
}
