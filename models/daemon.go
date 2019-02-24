package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DaemonJob struct {
	gorm.Model
	Name            string    `json:"name" gorm:"unique;not null"`
	ErrorMailNotify bool      `json:"errorMailNotify"`
	ErrorAPINotify  bool      `json:"errorAPINotify"`
	Disabled        bool      `json:"disabled"`
	Status          JobStatus `json:"status"`
	MailTo          string    `json:"mailTo"`
	APITo           string    `json:"APITo"`
	FailRestart     bool      `json:"failRestart"`
	StartAt         time.Time `json:"startAt"`
	User            string    `json:"user"`
	WorkDir         string    `json:"workDir"`

	Commands StringSlice `json:"commands" gorm:"type:TEXT"`
}
