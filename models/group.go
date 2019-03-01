package models

import "github.com/jinzhu/gorm"

type Group struct {
	gorm.Model
	Name string `json:"name" gorm:"not null; unique"`
}

func (g *Group) Save() error {
	return DB().Save(g).Error
}
