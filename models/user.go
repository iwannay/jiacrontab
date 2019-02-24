package models

import (
	"crypto/md5"
	"fmt"
	"github.com/iwannay/log"
	"jiacrontab/pkg/util"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"not null; unique"`
	Passwd   string `json:"passwd"`
	Salt     string `json:"salt"`
	GroupID  uint   `json:"groupID"`
	Root     bool   `json:"root"`
	Mail     string `json:"mail"`
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
func (u *User) Verify(username, passwd string) bool {
	ret := DB().Take(u, "username=?", username)

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

func (u *User) setPasswd() {
	u.Salt = u.getSalt()
	bts := md5.Sum([]byte(fmt.Sprint(u.Passwd, u.Salt)))
	u.Passwd = fmt.Sprintf("%x", bts)
}

func (u *User) Create() error {
	u.setPasswd()
	return DB().Create(u).Error
}

func (u *User) SetGroup() error {
	return DB().Model(u).Where("user_id=?", u.ID).Update("group_id", u.GroupID).Error
}
