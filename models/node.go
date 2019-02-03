package models

import "github.com/jinzhu/gorm"

type Node struct {
	gorm.Model
	Name string `gorm:"not null"`
	Info string
	Addr string `gorm:"unique;not null"`
}

func (c *Node) Delete(addr string) error {
	ret := DB().Unscoped().Delete(c, "addr=?", addr)
	return ret.Error
}
