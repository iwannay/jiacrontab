package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

type CrontabJob struct {
	gorm.Model
	Name            string `gorm:"unique;not null"`
	Commands        []string
	DependJobs      DependJobs   `gorm:"type:TEXT"`
	PipeCommands    PipeComamnds `gorm:"type:TEXT"`
	LastCostTime    int64
	LastExecTime    time.Time
	NextExecTime    time.Time
	LastExitStatus  string
	Timeout         int64
	ErrorMailNotify bool
	ErrorAPINotify  bool
	IsSync          bool // 脚本是否同步执行
	MailTo          string
	ApiTo           string
	MaxConcurrent   int      // 脚本最大并发量
	TimeoutTrigger  string   // email/kill/email_and_kill/ignore/api
	TimeArgs        TimeArgs `gorm:"type:TEXT"`
}

type DependJobs []DependJob

func (d *DependJobs) Scan(v interface{}) error {
	switch val := v.(type) {
	case string:
		return json.Unmarshal([]byte(val), d)
	case []byte:
		return json.Unmarshal(val, d)
	default:
		return errors.New("not support")
	}

}

func (d DependJobs) Value() (driver.Value, error) {
	bts, err := json.Marshal(d)
	return string(bts), err
}

type TimeArgs struct {
	Weekday string
	Month   string
	Day     string
	Hour    string
	Minute  string
}

func (c *TimeArgs) Scan(v interface{}) error {
	switch val := v.(type) {
	case string:
		return json.Unmarshal([]byte(val), c)
	case []byte:
		return json.Unmarshal(val, c)
	default:
		return errors.New("not support")
	}

}

func (c TimeArgs) Value() (driver.Value, error) {
	bts, err := json.Marshal(c)
	return string(bts), err
}

type DependJob struct {
	Name     string
	Dest     string
	From     string
	JobID    int
	ID       int
	Commands []string
	Timeout  int64
	Err      string
}

type PipeComamnds [][]string

func (p *PipeComamnds) Scan(v interface{}) error {
	switch val := v.(type) {
	case string:
		return json.Unmarshal([]byte(val), p)
	case []byte:
		return json.Unmarshal(val, p)
	default:
		return errors.New("not support")
	}

}

func (p PipeComamnds) Value() (driver.Value, error) {
	bts, err := json.Marshal(p)
	return string(bts), err
}
