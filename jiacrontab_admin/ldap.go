package admin

import (
	"fmt"
	"jiacrontab/models"
	"time"

	"errors"

	ld "github.com/go-ldap/ldap/v3"
)

type Ldap struct {
	BindUserDn             string
	BindPwd                string
	BaseOn                 string
	UserField              string
	Addr                   string
	Timeout                time.Duration
	fields                 []map[string]string
	queryFields            []string
	lastSynced             time.Time
	DisabledAnonymousQuery bool
}

func (l *Ldap) connect() (*ld.Conn, error) {
	var err error
	conn, err := ld.DialURL(l.Addr)
	if err != nil {
		return nil, err
	}
	conn.SetTimeout(l.Timeout)
	return conn, nil
}

func (l *Ldap) Login(username string, password string) (*models.User, error) {
	err := l.loadLdapFields()
	if err != nil {
		return nil, err
	}

	conn, err := l.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if l.DisabledAnonymousQuery {
		err := conn.Bind(l.BindUserDn, l.BindPwd)
		if err != nil {
			return nil, err
		}
	}
	query := ld.NewSearchRequest(
		l.BaseOn,
		ld.ScopeWholeSubtree,
		ld.DerefAlways,
		0, 0, false,
		fmt.Sprintf("(%v=%v)", l.UserField, username),
		l.queryFields, nil,
	)
	ret, err := conn.Search(query)

	if err != nil {
		return nil, fmt.Errorf("在Ldap搜索用户失败 - %v", err)
	}
	if len(ret.Entries) == 0 {
		return nil, fmt.Errorf("在Ldap中未查询到对应的用户信息 - %v", err)
	}

	err = conn.Bind(ret.Entries[0].DN, password)
	if err != nil {
		return nil, errors.New("帐号或密码不正确")
	}
	return l.convert(ret.Entries[0])
}

func (l *Ldap) loadLdapFields() error {
	var setting models.SysSetting
	var list []map[string]string

	if time.Since(l.lastSynced).Hours() < 1 {
		return nil
	}

	err := models.DB().Model(&models.SysSetting{}).Where("class=?", 1).Find(&setting).Error
	if err != nil {
		return err
	}
	// err = json.Unmarshal(setting.Content, &list)
	// if err != nil {
	// 	return err
	// }
	for _, v := range list {
		if v["ldap_field_name"] != "" {
			l.fields = append(l.fields, v)
		}
	}

	l.queryFields = []string{"dn"}
	for _, v := range l.fields {
		l.queryFields = append(l.queryFields, v["ldap_field_name"])
	}

	return nil
}

func (l *Ldap) convert(ldapUserInfo *ld.Entry) (*models.User, error) {
	var userinfo models.User
	for _, v := range l.fields {
		switch v["local_field_name"] {
		case "username":
			userinfo.Username = ldapUserInfo.GetAttributeValue(v["ldap_field_name"])
		case "gender":
			userinfo.Gender = ldapUserInfo.GetAttributeValue(v["ldap_field_name"])
		case "avatar":
			userinfo.Avatar = ldapUserInfo.GetAttributeValue(v["ldap_field_name"])
		case "email":
			userinfo.Mail = ldapUserInfo.GetAttributeValue(v["ldap_field_name"])
		}
	}
	err := models.DB().Where("username=?", userinfo.Username).FirstOrCreate(&userinfo).Error
	return &userinfo, err
}
