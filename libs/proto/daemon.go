package proto

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DaemonTask struct {
	gorm.Model
	ID         int64
	Name       string `gorm:"unique;not null"`
	MailNofity bool
	Status     int
	MailTo     string
	Command    string
	Args       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time
}
