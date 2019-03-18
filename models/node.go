package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type Node struct {
	gorm.Model
	Name           string `json:"name" gorm:"not null"`
	DaemonTaskNum  int    `json:"daemonTaskNum"`
	Disabled       bool   `json:"disabled"` // 通信失败时Disabled会被设置为true
	CrontabTaskNum int    `json:"crontabTaskNum"`
	GroupID        uint   `json:"groupID" gorm:"not null;unique_index:uni_group_addr" `
	Addr           string `json:"addr"gorm:"not null;unique_index:uni_group_addr"`
}

func (n *Node) VerifyUserGroup(userID, groupID uint, addr string) bool {
	var user User
	if DB().Take(&user, "id=? and group_id=?", userID, groupID).Error != nil {
		return false
	}

	if groupID == 0 {
		return true
	}

	return n.Exists(groupID, addr)
}

func (n *Node) Delete(groupID uint, addr string) error {
	ret := DB().Debug().Delete(n, "group_id=? and addr=?", groupID, addr)

	if ret.Error != nil {
		return ret.Error
	}

	if ret.RowsAffected == 0 {
		return errors.New("Delete failed")
	}
	return nil
}

func (n *Node) Rename(groupID uint, addr string) error {
	return DB().Model(n).Where("group_id=? and addr=?", groupID, addr).Updates(n).Error
}

// GroupNode 为节点分组，复制groupID=0分组中node至目标分组
func (n *Node) GroupNode(addr string, targetGroupID uint, targetNodeName, targetGroupName string) error {

	if targetGroupID == 0 {
		group := &Group{
			Name: targetGroupName,
		}
		if err := DB().Save(group).Error; err != nil {
			return err
		}
		targetGroupID = group.ID
	}

	err := DB().Model(n).Debug().Where("group_id=? and addr=?", 0, addr).Take(n).Error
	if err != nil {
		return err
	}

	if targetNodeName == "" {
		targetNodeName = n.Name
	}

	return DB().Save(&Node{
		Addr:    addr,
		GroupID: targetGroupID,
		Name:    targetNodeName,
	}).Error
}

func (n *Node) Exists(groupID uint, addr string) bool {
	if DB().Take(n, "group_id=? and addr=?", groupID, addr).Error != nil {
		return false
	}
	return true
}
