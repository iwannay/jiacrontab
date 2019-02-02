package jiacrontabd

import (
	"fmt"
	"jiacrontab/pkg/version"
	"log"
	"os"
	"reflect"
	"strings"

	ini "gopkg.in/ini.v1"
)

const (
	configFile = "client.ini"
	appname    = "jiacrontabd"
)

var cfg *Config

type Config struct {
	LogLevel         string `opt:"log_level" default:"warn"`
	VerboseJobLog    bool   `opt:"verbose_job_log" default:"true"`
	ListenAddr       string `opt:"listen_addr" default:":20001"`
	LocalAddr        string `opt:"local_addr default:"127.0.0.1:20002"`
	AdminAddr        string `opt:"admin_addr" default:"127.0.0.1:20003"`
	LogPath          string `opt:"log_path" default:"./logs"`
	PprofAddr        string `opt:"pprof_addr" default:"127.0.0.1:20004"`
	MailTo           string `opt:"mail_to" default:""`
	AutoCleanTaskLog bool   `opt:"auto_clean_task_log" default:"true"`
	ClientID         string `opt:"clent_id" default:""`
	Hostname         string `opt:"hostname" default:""`
	UserAgent        string `opt:"user_agent" default:""`
}

func (c *Config) init() error {
	cf := loadConfig()
	val := reflect.ValueOf(c).Elem()
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		opt := field.Tag.Get("opt")

		defaultVal := cf.Section("jiacrontabd").Key(opt).String()

		if defaultVal == "" {
			defaultVal = field.Tag.Get("default")
		}

		if defaultVal == "" || opt == "" {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Bool:
			setVal := false
			if defaultVal == "true" {
				setVal = true
			}
			val.Field(i).SetBool(setVal)
		case reflect.String:
			val.Field(i).SetString(defaultVal)
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("ERROR: unable to get hostname %s", err.Error())
	}

	c.ClientID = strings.Split(hostname, ".")[0]
	c.Hostname = hostname
	c.UserAgent = fmt.Sprintf("%s", version.String(appname))
	return nil
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
