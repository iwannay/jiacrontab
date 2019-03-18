package models

import (
	"github.com/jinzhu/gorm"
)

type Group struct {
	gorm.Model
	Name string `json:"name" gorm:"not null; unique"`
}

func (g *Group) Save() error {
	if g.ID == 0 {
		return DB().Create(g).Error
	}
	return DB().Save(g).Error
}
