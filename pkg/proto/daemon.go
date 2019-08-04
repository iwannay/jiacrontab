package proto

import (
	"jiacrontab/models"
)

type EditDaemonJobArgs struct {
	Job     models.DaemonJob
	GroupID uint
	UserID  uint
	Root    bool
}
