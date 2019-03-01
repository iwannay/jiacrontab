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

func (n *Node) VerifyUserGroup(userID, groupID uint, addr string) bool {
	var user User
	if DB().Take(user, "user_id=? and group_id", userID, groupID).Error != nil {
		return false
	}
	return n.Exists(groupID, addr)
}

func (n *Node) Delete(groupID uint, addr string) error {
	return DB().Delete(n, "disabled=1 and group_id=? and addr=?", groupID, addr).Error
}

func (n *Node) Rename(groupID uint, addr string) error {
	return DB().Model(n).Where("group_id=? and addr=?", groupID, addr).Updates(n).Error
}

func (n *Node) Exists(groupID uint, addr string) bool {
	if DB().Take(n, "group_id=? and addr=?", groupID, addr).Error != nil {
		return false
	}
	return true
}

func (n *Node) SetGroup() error {
	n.ID = 0
	return DB().Where("group_id=? and addr =?", n.GroupID, n.Addr).Save(n).Error
}
