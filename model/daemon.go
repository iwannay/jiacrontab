package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DaemonTask struct {
	gorm.Model
	Name          string `gorm:"unique;not null"`
	MailNotify    bool
	Status        int
	MailTo        string
	FailedRestart bool
	ProcessNum    int
	StartAt       time.Time
	Command       string
	Args          string
}
