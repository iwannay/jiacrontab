package admin

import (
	"errors"
	"jiacrontab/pkg/mailer"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/util"
	"reflect"
	"time"

	"github.com/kataras/iris"

	ini "gopkg.in/ini.v1"
)

const (
	configFile = "jiacrontab_admin.ini"
	appname    = "jiacrontab"
)

var cfg *Config

type appOpt struct {
	HTTPListenAddr string `opt:"http_listen_addr" default:":20000"`
	StaticDir      string `opt:"static_dir" default:"./static"`
	TplDir         string `opt:"tpl_dir" default:"tpl_dir"`
	TplExt         string `opt:"tpl_ext" default:"tpl_ext"`
	RPCListenAddr  string `opt:"rpc_listen_addr" default:":20003"`
	AppName        string `opt:"app_name" default:"jiacrontab"`
	FirstUse       bool   `opt:"first_use" default:"true"`
	SigningKey     string `opt:"signing_key" default:"WERRTT1234$@#@@$"`
}

type jwtOpt struct {
	SigningKey string `opt:"signing_key" default:"ADSFdfs2342$@@#"`
	Name       string `opt:"name" default:"token"`
	Expires    int64  `opt:"expires" default:"3600"`
}

type mailerOpt struct {
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
	DriverName string `opt:"driver_name" default:"sqlite3"`
	DSN        string `opt:"dsn" default:"data/jiacrontab_admin.db"`
}

type Config struct {
	Mailer          *mailerOpt   `section:"mail"`
	Jwt             *jwtOpt      `section:"jwt"`
	App             *appOpt      `section:"app"`
	Database        *databaseOpt `section:"database"`
	iniFile         *ini.File
	ServerStartTime time.Time `json:"-"`
}

func (c *Config) SetUsed() {
	c.App.FirstUse = false
	c.iniFile.Section("app").NewKey("first_use", "false")
	c.iniFile.SaveTo(configFile)
}

func (c *Config) init() {
	c.iniFile = loadConfig()
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

			if defaultVal == "" {
				defaultVal = subField.Tag.Get("default")
			}
			if defaultVal == "" {
				continue
			}

			switch subField.Type.Kind() {
			case reflect.Bool:
				setVal := false
				if defaultVal == "true" {
					setVal = true
				}
				subVal.Field(j).SetBool(setVal)
			case reflect.String:
				subVal.Field(j).SetString(defaultVal)
			case reflect.Int64:
				subVal.Field(j).SetInt(util.ParseInt64(defaultVal))
			}
			key.SetValue(defaultVal)
		}
	}
}

func newConfig() *Config {
	c := &Config{
		App:      &appOpt{},
		Mailer:   &mailerOpt{},
		Jwt:      &jwtOpt{},
		Database: &databaseOpt{},
	}
	c.init()
	return c
}

func loadConfig() *ini.File {
	f, err := ini.Load(configFile)
	if err != nil {
		panic(err)
	}
	return f
}

func getConfig(c iris.Context) {

	ctx := wrapCtx(c)
	gid, err := ctx.getGroupIDFromToken()
	if err != nil || gid != 0 {
		ctx.respError(proto.Code_Error, "无权访问", nil)
		return
	}
	ctx.respSucc("", cfg)
}

func sendTestMail(c iris.Context) {
	var (
		ctx     = wrapCtx(c)
		err     error
		reqBody SendTestMailReqParams
	)

	if err = reqBody.verify(ctx); err != nil {
		goto failed
	}

	if cfg.Mailer.Enabled {
		err = mailer.SendMail([]string{reqBody.MailTo}, "测试邮件", "测试邮件请勿回复！")
		if err != nil {
			goto failed
		}
		ctx.respSucc("", nil)
		return
	}

	err = errors.New("邮箱服务未开启")

failed:
	ctx.respError(proto.Code_Error, err.Error(), nil)
}

func init() {
	cfg = newConfig()
}
