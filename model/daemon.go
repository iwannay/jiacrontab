package model

import (
	"github.com/jinzhu/gorm"
)

type DaemonTask struct {
	gorm.Model
	Name       string `gorm:"unique;not null"`
	MailNofity bool
	Status     int
	MailTo     string
	Command    string
	Args       string
}



