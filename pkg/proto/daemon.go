package proto

const (
	ActionStopDaemonJob = iota
	ActionStartDaemonJob
	ActionDeleteDaemonJob
)

type ActionDaemonJobArgs struct {
	Action int
	JobIDs []uint
}
