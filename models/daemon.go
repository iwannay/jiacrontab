package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DaemonJob struct {
	gorm.Model
	Name            string      `json:"name" gorm:"unique;not null"`
	ErrorMailNotify bool        `json:"errorMailNotify"`
	ErrorAPINotify  bool        `json:"errorAPINotify"`
	Disabled        bool        `json:"disabled"`
	MailTo          string      `json:"mailTo"`
	APITo           string      `json:"APITo"`
	Status          int         `json:"status"`
	FailRestart     bool        `json:"failRestart"`
	StartAt         time.Time   `json:"startAt"`
	Commands        StringSlice `json:"commands" gorm:"type:TEXT"`
}
