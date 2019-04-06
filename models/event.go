package models

import (
	"github.com/iwannay/log"

	"github.com/jinzhu/gorm"
)

type Event struct {
	gorm.Model
	GroupID    uint   `json:"groupID" gorm:"index"`
	Username   string `json:"username"`
	UserID     uint   `json:"userID"`
	EventDesc  string `json:"eventDesc"`
	TargetName string `json:"targetName"`
	NodeAddr   string `json:"nodeAddr" gorm:"index"`
	Content    string `json:"content"`
}

func (e *Event) Pub() {
	err := DB().Model(e).Create(e).Error
	if err != nil {
		log.Error("Event.Pub", err)
	}
}
