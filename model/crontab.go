package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/jinzhu/gorm"
)

type CrontabTask struct {
	gorm.Model
	Name               string `gorm:"unique;not null"`
	Command            string
	Depends            DependsTasks `gorm:"type:TEXT"`
	PipeCommands       PipeComamnds `gorm:"type:TEXT"`
	State              int          // 0/1/2
	Args               string
	Create             int64
	LastCostTime       int64
	LastExecTime       int64
	LastExitStatus     string
	Timeout            int64
	NumberProcess      int32
	TimerCounter       int32
	UnexpectedExitMail bool
	UnexpectedExitApi  bool
	Sync               bool // 脚本是否同步执行
	MailTo             string
	ApiTo              string
	MaxConcurrent      int         // 脚本最大并发量
	OpTimeout          string      // email/kill/email_and_kill/ignore/api
	C                  CrontabArgs `gorm:"type:TEXT"`
}

type DependsTasks []DependsTask

func (d *DependsTasks) Scan(v interface{}) error {
	switch val := v.(type) {
	case string:
		return json.Unmarshal([]byte(val), d)
	case []byte:
		return json.Unmarshal(val, d)
	default:
		return errors.New("not support")
	}

}

func (d DependsTasks) Value() (driver.Value, error) {
	bts, err := json.Marshal(d)
	return string(bts), err
}

type CrontabArgs struct {
	Weekday string
	Month   string
	Day     string
	Hour    string
	Minute  string
}

func (c *CrontabArgs) Scan(v interface{}) error {
	switch val := v.(type) {
	case string:
		return json.Unmarshal([]byte(val), c)
	case []byte:
		return json.Unmarshal(val, c)
	default:
		return errors.New("not support")
	}

}

func (c CrontabArgs) Value() (driver.Value, error) {
	bts, err := json.Marshal(c)
	return string(bts), err
}

type DependsTask struct {
	Name         string
	Dest         string
	From         string
	TaskId       uint
	Id           string `json:"-"`
	TaskEntityId string `json:"-"`
	Command      string
	Args         string
	Timeout      int64
	Err          string `json:"-"`
	LogContent   []byte `json:"-"`
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
