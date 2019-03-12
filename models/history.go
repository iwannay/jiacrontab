package models

import (
	"time"

	"github.com/iwannay/log"
	"github.com/jinzhu/gorm"
)

type JobHistory struct {
	gorm.Model
	JobType    uint8     `json:"jobType"` // 0:定时任务,1:常驻任务
	JobID      uint      `json:"jobID"`
	JobName    string    `json:"jobName"`
	ExitStatus uint      `json:"exitStatus"`
	StartTime  time.Time `json:"execTime"`
	EndTime    time.Time `json:"endTime"`
}

func PushJobHistory(job interface{}, endTime time.Time) {
	var err error
	switch job := job.(type) {
	case *DaemonJob:
		err = DB().Omit("updated_at", "created_at", "deleted_at").Save(&JobHistory{
			JobID:      job.ID,
			JobType:    1,
			JobName:    job.Name,
			ExitStatus: 0,
			StartTime:  job.StartAt,
		}).Error
	case *CrontabJob:
		err = DB().Omit("updated_at", "created_at", "deleted_at").Save(&JobHistory{
			JobID:      job.ID,
			JobType:    0,
			JobName:    job.Name,
			ExitStatus: 0,
			StartTime:  job.LastExecTime,
			EndTime:    endTime,
		}).Error
	}
	if err != nil {
		log.Error("PushJobHistory failed:", err)
	}
}
