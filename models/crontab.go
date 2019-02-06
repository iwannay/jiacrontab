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
	Name            string       `json:"name" gorm:"unique;not null"`
	Commands        StringSlice  `json:"commands" gorm:"type:TEXT"`
	DependJobs      DependJobs   `json:"dependJobs" gorm:"type:TEXT"`
	PipeCommands    PipeComamnds `json:"pipeCommands" gorm:"type:TEXT"`
	LastCostTime    int64        `json:"lastCostTime"`
	LastExecTime    time.Time    `json:"lastExecTime"`
	NextExecTime    time.Time    `json:"nextExecTime"`
	LastExitStatus  string       `json:"lastExitStatus"`
	Timeout         int          `json:"timeout"`
	ErrorMailNotify bool         `json:"errorMailNotify"`
	ErrorAPINotify  bool         `json:"errorAPINotify"`
	IsSync          bool         `json:"isSync"` // 脚本是否同步执行
	MailTo          string       `json:"mailTo"`
	APITo           string       `json:"APITo"`
	MaxConcurrent   uint         `json:"maxConcurrent"`  // 脚本最大并发量
	TimeoutTrigger  string       `json:"timeoutTrigger"` // email/kill/email_and_kill/ignore/api
	TimeArgs        TimeArgs     `json:"timeArgs" gorm:"type:TEXT"`
}
type StringSlice []string

func (s *StringSlice) Scan(v interface{}) error {
	switch val := v.(type) {
	case string:
		return json.Unmarshal([]byte(val), s)
	case []byte:
		return json.Unmarshal(val, s)
	default:
		return errors.New("not support")
	}

}

func (s StringSlice) Value() (driver.Value, error) {
	bts, err := json.Marshal(s)
	return string(bts), err
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
	Weekday string `json:"weekday"`
	Month   string `json:"month"`
	Day     string `json:"day"`
	Hour    string `json:"hour"`
	Minute  string `json:"minute"`
	Second  string `json:"second"`
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
	Name     string   `json:"name"`
	Dest     string   `json:"dest"`
	From     string   `json:"from"`
	JobID    int      `json:"jobID"`
	ID       string   `json:"id"`
	Commands []string `json:"commands"`
	Timeout  int64    `json:"timeout"`
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
