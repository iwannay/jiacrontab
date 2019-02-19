package models

import "github.com/jinzhu/gorm"

type Node struct {
	gorm.Model
	Name           string `json:"name" gorm:"not null"`
	DaemonTaskNum  int    `json:"daemonTaskNum"`
	Disabled       bool   `json:"disabled"`
	CrontabTaskNum int    `json:"crontabTaskNum"`
	GroupID        uint   `json:"groupID" gorm:"not null;unique_index:uni_group_addr" `
	Addr           string `json:"addr"gorm:"not null;unique_index:uni_group_addr"`
}

func (n *Node) Delete(id int) error {
	return DB().Delete(n, "id=?", id).Error

}

func (n *Node) Rename() error {
	return DB().Where("id=? and addr=?", n.ID, n.Addr).Updates(n).Error
}

func (n *Node) Exists() bool {
	if DB().Take(n, "group_id=? and addr=?", n.GroupID, n.Addr).Error != nil {
		return false
	}
	return true
}

func (n *Node) SetGroup() error {
	n.ID = 0
	return DB().Where("group_id=? and addr =?", n.GroupID, n.Addr).Save(n).Error
}
