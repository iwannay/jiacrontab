package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DaemonJob struct {
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
	Commands      []string
}
