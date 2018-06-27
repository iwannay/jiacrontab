package model

import "github.com/jinzhu/gorm"

type Client struct {
	gorm.Model
	Name   string `gorm:"unique;not null"`
	Status int
	Addr   string `gorm:"unique;not null"`
	Mail   string
}
