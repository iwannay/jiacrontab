package proto

import (
	"jiacrontab/models"
)

type EditDaemonJobArgs struct {
	Job  models.DaemonJob
	Root bool
}
