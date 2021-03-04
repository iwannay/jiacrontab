package models

import (
	"time"

	"gorm.io/gorm"
)

type DaemonJob struct {
	gorm.Model
	Name            string      `json:"name" gorm:"index;not null"`
	GroupID         uint        `json:"groupID" grom:"index"`
	Command         StringSlice `json:"command" gorm:"type:varchar(1000)"`
	Code            string      `json:"code"  gorm:"type:TEXT"`
	ErrorMailNotify bool        `json:"errorMailNotify"`
	ErrorAPINotify  bool        `json:"errorAPINotify"`
	ErrorDingdingNotify  bool   `json:"errorDingdingNotify"`
	Status          JobStatus   `json:"status"`
	MailTo          StringSlice `json:"mailTo" gorm:"type:varchar(1000)"`
	APITo           StringSlice `json:"APITo" gorm:"type:varchar(1000)"`
	DingdingTo      StringSlice `json:"DingdingTo" gorm:"type:varchar(1000)"`
	FailRestart     bool        `json:"failRestart"`
	RetryNum        int         `json:"retryNum"`
	StartAt         time.Time   `json:"startAt"`
	WorkUser        string      `json:"workUser"`
	WorkIp          StringSlice `json:"workIp" gorm:"type:varchar(1000)"`
	WorkEnv         StringSlice `json:"workEnv" gorm:"type:varchar(1000)"`
	WorkDir         string      `json:"workDir"`
	CreatedUserID   uint        `json:"createdUserId"`
	CreatedUsername string      `json:"createdUsername"`
	UpdatedUserID   uint        `json:"updatedUserID"`
	UpdatedUsername string      `json:"updatedUsername"`
}
