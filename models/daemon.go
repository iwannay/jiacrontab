package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DaemonJob struct {
	gorm.Model
	Name            string `gorm:"unique;not null"`
	ErrorMailNotify bool
	ErrorAPINotify  bool
	Disabled        bool
	MailTo          string
	APITo           string
	FailRestart     bool
	StartAt         time.Time
	Commands        []string
}
