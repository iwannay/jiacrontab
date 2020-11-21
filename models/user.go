package models

import (
	"crypto/md5"
	"errors"
	"fmt"
	"jiacrontab/pkg/util"
	"time"

	"github.com/iwannay/log"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"not null; unique"`
	Passwd   string `json:"-"`
	Salt     string `json:"-"`
	Avatar   string `json:"avatar"`
	Version  int64  `json:"version"`
	GroupID  uint   `json:"groupID" grom:"index"`
	Root     bool   `json:"root"`
	Mail     string `json:"mail"`
	Group    Group  `json:"group"`
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

// Verify 验证用户
func (u *User) VerifyByUserId(id uint, passwd string) bool {
	ret := DB().Take(u, "id=?", id)

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
	if u.Passwd == "" {
		return
	}
	u.Salt = u.getSalt()
	bts := md5.Sum([]byte(fmt.Sprint(u.Passwd, u.Salt)))
	u.Passwd = fmt.Sprintf("%x", bts)
}

func (u *User) Create() error {
	u.setPasswd()
	u.Version = time.Now().Unix()
	return DB().Create(u).Error
}

func (u User) Update() error {
	u.setPasswd()
	u.Version = time.Now().Unix()
	return DB().Model(&u).Updates(u).Error
}

func (u *User) Delete() error {
	if err := DB().Take(u, "id=?", u.ID).Error; err != nil {
		return err
	}
	return DB().Delete(u).Error
}

func (u *User) SetGroup(group *Group) error {

	if u.GroupID != 0 {
		if err := DB().Take(group, "id=?", u.GroupID).Error; err != nil {
			return fmt.Errorf("查询分组失败：%s", err)
		}
	}
	if u.ID == 1 {
		return errors.New("系统用户不允许修改")
	}

	defer DB().Take(u, "id=?", u.ID)

	return DB().Model(u).Where("id=?", u.ID).Updates(map[string]interface{}{
		"group_id": u.GroupID,
		"version":  time.Now().Unix(),
		"root":     u.Root,
	}).Error
}
