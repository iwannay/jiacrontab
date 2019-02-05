package models

import "github.com/jinzhu/gorm"

type Node struct {
	gorm.Model
	Name           string `gorm:"not null"`
	Disabled       bool
	DaemonTaskNum  int
	CrontabTaskNum int
	Mail           string
	GroupID        int    `gorm:"not null;unique_index:uni_group_addr" `
	Addr           string `gorm:"not null;unique_index:uni_group_addr"`
}

func (c *Node) Delete(id int) error {
	ret := DB().Unscoped().Delete(c, "id=?", id)
	return ret.Error
}
