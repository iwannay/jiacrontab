package models

import "github.com/jinzhu/gorm"

type Node struct {
	gorm.Model
	Name           string `json:"name" gorm:"not null"`
	Disabled       bool   `json:"disabled"`
	DaemonTaskNum  int    `json:"daemonTaskNum"`
	CrontabTaskNum int    `json:"crontabTaskNum"`
	Mail           string `json:"mail"`
	GroupID        int    `json:"groupID" gorm:"not null;unique_index:uni_group_addr" `
	Addr           string `json:"addr"gorm:"not null;unique_index:uni_group_addr"`
}

func (c *Node) Delete(id int) error {
	ret := DB().Unscoped().Delete(c, "id=?", id)
	return ret.Error
}
