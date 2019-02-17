package models

import (
	"jiacrontab/pkg/log"

	"github.com/jinzhu/gorm"
)

type Event struct {
	gorm.Model
	GroupID   uint   `json:"group_id" gorm:"index"`
	Username  string `json:"username"`
	UserID    uint   `json:"user_id"`
	EventDesc string `json:"event_desc"`
	NodeAddr  string `json:"node_addr" gorm:"index"`
	Content   string `json:"content"`
}

func (e *Event) Pub() {
	err := DB().Model(e).Create(e).Error
	if err != nil {
		log.Error("Event.Pub", err)
	}
}
