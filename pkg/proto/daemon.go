package proto

const (
	StopDaemonTask = iota
	StartDaemonTask
	DeleteDaemonTask
)

type ActionDaemonJobArgs struct {
	Action int
	JobIDs []uint
}
