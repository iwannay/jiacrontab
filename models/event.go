package models

import (
	"github.com/iwannay/log"

	"gorm.io/gorm"
)

type EventSourceName string
type EventSourceUsername string

type Event struct {
	gorm.Model
	GroupID        uint   `json:"groupID" gorm:"index"`
	Username       string `json:"username"`
	UserID         uint   `json:"userID" gorm:"index"`
	EventDesc      string `json:"eventDesc"`
	TargetName     string `json:"targetName"`
	SourceUsername string `json:"sourceUsername"`
	SourceName     string `json:"sourceName" gorm:"index;size:500"`
	Content        string `json:"content"`
}

func (e *Event) Pub() {
	err := DB().Model(e).Create(e).Error
	if err != nil {
		log.Error("Event.Pub", err)
	}
}
