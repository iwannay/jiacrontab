package jiacrontabd

import (
	"fmt"
	"jiacrontab/pkg/util"
	"jiacrontab/pkg/version"
	"log"
	"reflect"

	ini "gopkg.in/ini.v1"
)

const (
	configFile = "jiacrontabd.ini"
	appname    = "jiacrontabd"
)

var cfg *Config

type Config struct {
	LogLevel         string `opt:"log_level" default:"warn"`
	VerboseJobLog    bool   `opt:"verbose_job_log" default:"true"`
	ListenAddr       string `opt:"listen_addr" default:":20001"`
	LocalAddr        string `opt:"local_addr" default:"127.0.0.1:20002"`
	AdminAddr        string `opt:"admin_addr" default:"127.0.0.1:20003"`
	LogPath          string `opt:"log_path" default:"./logs"`
	PprofAddr        string `opt:"pprof_addr" default:"127.0.0.1:20004"`
	MailTo           string `opt:"mail_to" default:""`
	AutoCleanTaskLog bool   `opt:"auto_clean_task_log" default:"true"`
	Hostname         string `opt:"hostname" default:""`
	UserAgent        string `opt:"user_agent" default:""`
	iniFile          *ini.File
	FirstUse         bool   `opt:"first_use" default:"true"`
	DriverName       string `opt:"driver_name" default:"sqlite3"`
	DSN              string `opt:"dsn" default:"data/jiacrontab_admin.db"`
}

func (c *Config) SetUsed() {
	c.FirstUse = false
	c.iniFile.Section("jiacrontabd").NewKey("first_use", "false")
	c.iniFile.SaveTo(configFile)
}

func (c *Config) init() error {
	c.iniFile = loadConfig()
	val := reflect.ValueOf(c).Elem()
	typ := val.Type()
	hostname := util.GetHostname()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		opt := field.Tag.Get("opt")
		if opt == "" {
			continue
		}
		key := c.iniFile.Section("jiacrontabd").Key(opt)
		defaultVal := key.String()

		if defaultVal == "" {
			defaultVal = field.Tag.Get("default")
		}

		if defaultVal == "" {
			switch opt {
			case "hostname":
				val.Field(i).SetString(hostname)
				key.SetValue(hostname)
			case "user_agent":
				ua := fmt.Sprintf("%s/%s", hostname, version.String(appname))
				val.Field(i).SetString(ua)
				key.SetValue(ua)
			}
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
		key.SetValue(defaultVal)
	}

	c.SetUsed()
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
	log.Printf("%+v", cfg)
}
