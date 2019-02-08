package proto

const (
	ActionStopDaemonTask = iota
	ActionStartDaemonTask
	ActionDeleteDaemonTask
)

type ActionDaemonJobArgs struct {
	Action int
	JobIDs []uint
}
