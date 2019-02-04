package models

import "github.com/jinzhu/gorm"

type Group struct {
	gorm.Model
	Name     string `gorm:"not null; unique"`
	NodeAddr string `gorm:"not null; unique"`
}
