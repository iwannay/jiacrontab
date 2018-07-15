package model

import "github.com/jinzhu/gorm"

type Client struct {
	gorm.Model
	Name           string `gorm:"not null"`
	State          int
	DaemonTaskNum  int
	CrontabTaskNum int
	Addr           string `gorm:"unique;not null"`
	Mail           string
}
