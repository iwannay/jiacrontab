package admin

import (
	"errors"
	"fmt"
	"jiacrontab/models"
	"jiacrontab/pkg/mailer"
	"jiacrontab/pkg/util"
	"reflect"
	"time"

	"os"

	ini "gopkg.in/ini.v1"
)

const (
	appname = "jiacrontab"
)

type AppOpt struct {
	HTTPListenAddr string `opt:"http_listen_addr" default:":20000"`
	RPCListenAddr  string `opt:"rpc_listen_addr" default:":20003"`
	AppName        string `opt:"app_name" default:"jiacrontab"`
	SigningKey     string `opt:"signing_key" default:"WERRTT1234$@#@@$"`
}

type JwtOpt struct {
	SigningKey string `opt:"signing_key" default:"ADSFdfs2342$@@#"`
	Name       string `opt:"name" default:"token"`
	Expires    int64  `opt:"expires" default:"3600"`
}

type MailerOpt struct {
	Enabled        bool   `opt:"enabled" default:"true"`
	QueueLength    int    `opt:"queue_length" default:"1000"`
	SubjectPrefix  string `opt:"subject_Prefix" default:"jiacrontab"`
	Host           string `opt:"host" default:""`
	From           string `opt:"from" default:"jiacrontab"`
	FromEmail      string `opt:"from_email" default:"jiacrontab"`
	User           string `opt:"user" default:"user"`
	Passwd         string `opt:"passwd" default:"passwd"`
	DisableHelo    bool   `opt:"disable_helo" default:""`
	HeloHostname   string `opt:"helo_hostname" default:""`
	SkipVerify     bool   `opt:"skip_verify" default:"true"`
	UseCertificate bool   `opt:"use_certificate" default:"false"`
	CertFile       string `opt:"cert_file" default:""`
	KeyFile        string `opt:"key_file" default:""`
	UsePlainText   bool   `opt:"use_plain_text" default:""`
}

type databaseOpt struct {
	DriverName string `opt:"driver_name"`
	DSN        string `opt:"dsn"`
}

type Config struct {
	Mailer          *MailerOpt   `section:"mail"`
	Jwt             *JwtOpt      `section:"jwt"`
	App             *AppOpt      `section:"app"`
	Database        *databaseOpt `section:"database"`
	CfgPath         string
	iniFile         *ini.File
	ServerStartTime time.Time `json:"-"`
}

// SetUsed 设置app已经激活
func (c *Config) Activate(opt *databaseOpt) error {
	if opt.DriverName == "sqlite3" || opt.DriverName == "sqlite" {
		opt.DriverName = "sqlite3"
		opt.DSN = "data/jiacrontab_admin.db"
	}

	c.Database = opt

	err := models.InitModel(c.Database.DriverName, c.Database.DSN)
	if err != nil {
		return err
	}
	c.iniFile.Section("database").Key("driver_name").SetValue(opt.DriverName)
	c.iniFile.Section("database").Key("dsn").SetValue(opt.DSN)
	return c.iniFile.SaveTo(c.CfgPath)
}

func (c *Config) Resolve() {

	c.ServerStartTime = time.Now()
	c.iniFile = loadConfig(c.CfgPath)
	defer c.iniFile.SaveTo(c.CfgPath)

	val := reflect.ValueOf(c).Elem()
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		section := field.Tag.Get("section")
		if section == "" {
			continue
		}
		subVal := reflect.ValueOf(val.Field(i).Interface()).Elem()
		subtyp := subVal.Type()
		for j := 0; j < subtyp.NumField(); j++ {
			subField := subtyp.Field(j)
			subOpt := subField.Tag.Get("opt")
			if subOpt == "" {
				continue
			}
			key := c.iniFile.Section(section).Key(subOpt)
			defaultVal := key.String()
			comment := subField.Tag.Get("comment")
			key.Comment = comment

			switch subField.Type.Kind() {
			case reflect.Bool:
				setVal := false
				if defaultVal == "true" {
					setVal = true
				}
				if subVal.Field(j).Bool() == false {
					subVal.Field(j).SetBool(setVal)
				}
				key.SetValue(fmt.Sprint(subVal.Field(j).Bool()))
			case reflect.String:
				if subVal.Field(j).String() == "" {
					subVal.Field(j).SetString(defaultVal)
				}
				key.SetValue(subVal.Field(j).String())
			case reflect.Int64:
				if subVal.Field(j).Int() == 0 {
					subVal.Field(j).SetInt(util.ParseInt64(defaultVal))
				}
				key.SetValue(fmt.Sprint(subVal.Field(j).Int()))
			}

		}
	}
}

func NewConfig() *Config {
	return &Config{
		App: &AppOpt{
			HTTPListenAddr: ":20000",
			RPCListenAddr:  ":20003",
			AppName:        "jiacrontab",
			SigningKey:     "WERRTT1234$@#@@$",
		},
		Mailer: &MailerOpt{
			Enabled:        true,
			QueueLength:    1000,
			SubjectPrefix:  "jiacrontab",
			From:           "jiacrontab",
			FromEmail:      "jiacrontab",
			SkipVerify:     true,
			UseCertificate: false,
		},
		Jwt: &JwtOpt{
			SigningKey: "ADSFdfs2342$@@#",
			Name:       "token",
			Expires:    3600,
		},
		Database: &databaseOpt{},
	}
}

func loadConfig(path string) *ini.File {
	f, err := util.TryOpen(path, os.O_CREATE)
	if err != nil {
		panic(err)
	}
	f.Close()

	iniFile, err := ini.Load(path)
	if err != nil {
		panic(err)
	}
	return iniFile
}

func GetConfig(ctx *myctx) {
	cfg := ctx.adm.getOpts()
	if !ctx.isSuper() {
		ctx.respNotAllowed()
		return
	}
	ctx.respSucc("", map[string]interface{}{
		"mail": map[string]interface{}{
			"host":            cfg.Mailer.Host,
			"user":            cfg.Mailer.User,
			"use_certificate": cfg.Mailer.UseCertificate,
			"skip_verify":     cfg.Mailer.SkipVerify,
			"cert_file":       cfg.Mailer.CertFile,
			"key_file":        cfg.Mailer.KeyFile,
		},
	})
}

func SendTestMail(ctx *myctx) {
	var (
		err     error
		reqBody SendTestMailReqParams
		cfg     = ctx.adm.getOpts()
	)

	if err = ctx.Valid(&reqBody); err != nil {
		ctx.respParamError(err)
		return
	}

	if cfg.Mailer.Enabled {
		err = mailer.SendMail([]string{reqBody.MailTo}, "测试邮件", "测试邮件请勿回复！")
		if err != nil {
			ctx.respBasicError(err)
			return
		}
		ctx.respSucc("", nil)
		return
	}

	ctx.respBasicError(errors.New("邮箱服务未开启"))
}
