package models

import (
	"crypto/md5"
	"fmt"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/util"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Name    string `gorm:"not null; unique"`
	Passwd  string
	Salt    string
	GroupID int
	Root    bool
	Mail    string
}

func (u *User) getSalt() string {
	var (
		seed = "1234567890!@#$%^&*()ABCDEFGHIJK"
		salt [10]byte
	)
	for k := range salt {
		salt[k] = seed[util.RandIntn(len(seed))]
	}

	return string(salt[0:10])
}

// Verify 验证用户
func (u *User) Verify(name, passwd string) bool {
	ret := DB().Take(u, "Name=?", name)

	if ret.Error != nil {
		log.Error("user.Verify:", ret.Error)
		return false
	}

	bts := md5.Sum([]byte(fmt.Sprint(passwd, u.Salt)))

	if fmt.Sprintf("%x", bts) == u.Passwd {
		return true
	}

	return false
}

func (u *User) Add() error {
	u.Salt = u.getSalt()
	return DB().Create(u).Error
}

func (u *User) SignUp() error {
	return DB().Create(u).Error
}
