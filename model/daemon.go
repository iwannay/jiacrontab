package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DaemonTask struct {
	gorm.Model
	Name       string `gorm:"unique;not null"`
	MailNofity bool
	Status     int
	MailTo     string
	StartTime  time.Time
	Command    string
	Args       string
}
