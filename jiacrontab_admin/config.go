package admin

import (
	"errors"
	"jiacrontab/pkg/file"
	"jiacrontab/pkg/mailer"
	"reflect"
	"time"

	"github.com/iwannay/log"

	ini "gopkg.in/ini.v1"
)

const (
	appname = "jiacrontab"
)

type AppOpt struct {
	HTTPListenAddr string `opt:"http_listen_addr"`
	RPCListenAddr  string `opt:"rpc_listen_addr"`
	AppName        string `opt:"app_name" `
	Debug          bool   `opt:"debug" `
	LogLevel       string `opt:"log_level"`
	SigningKey     string `opt:"signing_key"`
}

type JwtOpt struct {
	SigningKey string `opt:"signing_key"`
	Name       string `opt:"name" `
	Expires    int64  `opt:"expires"`
}

type MailerOpt struct {
	Enabled        bool   `opt:"enabled"`
	QueueLength    int    `opt:"queue_length"`
	SubjectPrefix  string `opt:"subject_Prefix"`
	Host           string `opt:"host"`
	From           string `opt:"from"`
	FromEmail      string `opt:"from_email"`
	User           string `opt:"user"`
	Passwd         string `opt:"passwd"`
	DisableHelo    bool   `opt:"disable_helo"`
	HeloHostname   string `opt:"helo_hostname"`
	SkipVerify     bool   `opt:"skip_verify"`
	UseCertificate bool   `opt:"use_certificate"`
	CertFile       string `opt:"cert_file"`
	KeyFile        string `opt:"key_file"`
	UsePlainText   bool   `opt:"use_plain_text"`
}

type SmnOpt struct {
	DomainName string `opt:"domainName"`
	UserName   string `opt:"userName"`
	UserPass   string `opt:"userPass"`
	Region     string `opt:"region"`
	TopicUrn   string `opt:"topicUrn"`

	CriticalTopicUrn   string `opt:"criticalTopicUrn"`
	ImportantTopicUrn   string `opt:"importantTopicUrn"`
	LessImportantTopicUrn   string `opt:"lessImportantTopicUrn"`
	InfoTopicUrn   string `opt:"infoTopicUrn"`
}

type databaseOpt struct {
	DriverName string `opt:"driver_name"`
	DSN        string `opt:"dsn"`
}

type Config struct {
	Mailer   *MailerOpt   `section:"mail"`
	Smn      *SmnOpt      `section:"smn"`

	Jwt      *JwtOpt      `section:"jwt"`
	App      *AppOpt      `section:"app"`
	Database *databaseOpt `section:"database"`

	CfgPath         string
	iniFile         *ini.File
	ServerStartTime time.Time `json:"-"`
}

func (c *Config) Resolve() {

	c.ServerStartTime = time.Now()
	c.iniFile = c.loadConfig(c.CfgPath)

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
			sec := c.iniFile.Section(section)

			if !sec.HasKey(subOpt) {
				continue
			}

			key := sec.Key(subOpt)

			switch subField.Type.Kind() {
			case reflect.Bool:
				v, err := key.Bool()
				if err != nil {
					log.Error(err)
					continue
				}
				subVal.Field(j).SetBool(v)
			case reflect.String:
				subVal.Field(j).SetString(key.String())
			case reflect.Int64:
				v, err := key.Int64()
				if err != nil {
					log.Error(err)
					continue
				}
				subVal.Field(j).SetInt(v)
			}
		}
	}
}

func NewConfig() *Config {
	return &Config{
		App: &AppOpt{
			Debug:          false,
			HTTPListenAddr: ":20000",
			RPCListenAddr:  ":20003",
			AppName:        "jiacrontab",
			SigningKey:     "WERRTT1234$@#@@$",
		},
		Mailer: &MailerOpt{
			Enabled:        false,
			QueueLength:    1000,
			SubjectPrefix:  "jiacrontab",
			SkipVerify:     true,
			UseCertificate: false,
		},
		Smn: &SmnOpt{
			DomainName: "",
			UserName: "",
			UserPass: "",
			Region: "",
			TopicUrn: "",
		},

		Jwt: &JwtOpt{
			SigningKey: "ADSFdfs2342$@@#",
			Name:       "token",
			Expires:    3600,
		},
		Database: &databaseOpt{},
	}
}

func (c *Config) loadConfig(path string) *ini.File {
	if !file.Exist(path) {
		f, err := file.CreateFile(path)
		if err != nil {
			panic(err)
		}
		f.Close()
	}

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
