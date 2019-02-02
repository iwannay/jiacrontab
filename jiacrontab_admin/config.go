package admin

import (
	"reflect"
	"time"

	ini "gopkg.in/ini.v1"
)

const (
	configFile = "client.ini"
	appname    = "jiacrontabd"
)

var cfg *Config

type appOpt struct {
	HttpListenAddr string `opt:"http_listen_addr" default:"20000"`
	StaticDir      string `opt:"static_dir" default:"./static"`
	TplDir         string `opt:"tpl_dir" default:"tpl_dir"`
	TplExt         string `opt:"tpl_ext" default:"tpl_ext"`
	RpcListenAddr  string `opt:"rpc_listen_addr" default:"20003"`

	AppName string `opt:"app_name" default:"jiacrontab"`
}

type jwtOpt struct {
	SigningKey string `opt:"signing_key" default:"ADSFdfs2342$@@#"`
	Name       string `opt:"name" default:"token"`
	Expires    int64  `opt:"name" default:"3600"`
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

type Config struct {
	Mailer          *mailerOpt `section:"mail"`
	Jwt             *jwtOpt    `section:"jwt`
	App             *appOpt    `section:"app"`
	ServerStartTime time.Time
}

func (c *Config) init() {
	cf := loadConfig()
	val := reflect.ValueOf(c).Elem()
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		opt := field.Tag.Get("opt")
		if opt == "" {
			continue
		}

		subVal := reflect.ValueOf(val.Field(i)).Elem()
		subtyp := subVal.Type()
		for j := 0; j < subtyp.NumField(); i++ {
			subField := typ.Field(i)
			subOpt := subField.Tag.Get("opt")
			defaultVal := cf.Section(opt).Key(subOpt).String()
			if defaultVal == "" {
				defaultVal = field.Tag.Get("default")
			}
			if defaultVal == "" || opt == "" {
				continue
			}

			switch subField.Type.Kind() {
			case reflect.Bool:
				setVal := false
				if defaultVal == "true" {
					setVal = true
				}
				subVal.Field(i).SetBool(setVal)
			case reflect.String:
				subVal.Field(i).SetString(defaultVal)
			}
		}
	}
}

func newConfig() *Config {
	c := &Config{}
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

func init() {
	cfg = newConfig()
}
