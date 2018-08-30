package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DaemonTask struct {
	gorm.Model
	Name          string `gorm:"unique;not null"`
	MailNotify    bool
	ApiNotify     bool
	Status        int
	MailTo        string
	ApiTo         string
	FailedRestart bool
	ProcessNum    int
	StartAt       time.Time
	Command       string
	Args          string
}
