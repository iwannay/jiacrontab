package models

import "github.com/jinzhu/gorm"

type Group struct {
	gorm.Model
	Name     string `json:"name" gorm:"not null; unique"`
	NodeAddr string `json:"nodeAddr" gorm:"not null; unique"`
}
