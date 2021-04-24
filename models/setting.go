package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type SysSetting struct {
	gorm.Model
	Class   int             `json:"class"`                                    // 设置分类，1 Ldap配置
	Content json.RawMessage `json:"content" gorm:"column:content; type:json"` // 配置内容
}
