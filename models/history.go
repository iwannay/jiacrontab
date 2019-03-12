package models

import (
	"time"

	"github.com/iwannay/log"
	"github.com/jinzhu/gorm"
)

const (
	JobTypeCrontab JobType = 0
	JobTypeDaemon  JobType = 1
)

type JobType uint8

type JobHistory struct {
	gorm.Model
	JobType   JobType   `json:"jobType"` // 0:定时任务,1:常驻任务
	JobID     uint      `json:"jobID"`
	JobName   string    `json:"jobName"`
	Addr      string    `json:"addr" gorm:"index"`
	ExitMsg   string    `json:"exitMsg"`
	StartTime time.Time `json:"execTime"`
	EndTime   time.Time `json:"endTime"`
}

func PushJobHistory(job *JobHistory) {
	err := DB().Create(job).Error
	if err != nil {
		log.Error("PushJobHistory failed:", err)
	}
}
