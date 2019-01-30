package proto

const (
	StopDaemonTask = iota
	StartDaemonTask
	DeleteDaemonTask
)

type ActionDaemonTaskArgs struct {
	Action  int
	TaskIds string
}
