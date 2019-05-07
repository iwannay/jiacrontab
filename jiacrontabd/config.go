package jiacrontabd

import (
	"fmt"
	"github.com/iwannay/log"
	"jiacrontab/pkg/file"
	"jiacrontab/pkg/util"
	"reflect"

	ini "gopkg.in/ini.v1"
)

const (
	appname = "jiacrontabd"
)

// var cfg *Config

type Config struct {
	LogLevel         string `opt:"log_level" default:"warn" comment:"日志等级"`
	VerboseJobLog    bool   `opt:"verbose_job_log" default:"true"`
	ListenAddr       string `opt:"listen_addr" default:":20001"`
	LocalAddr        string `opt:"local_addr" default:"127.0.0.1:20002"`
	AdminAddr        string `opt:"admin_addr" default:"127.0.0.1:20003"`
	LogPath          string `opt:"log_path" default:"./logs"`
	PprofAddr        string `opt:"pprof_addr" default:"127.0.0.1:20004"`
	AutoCleanTaskLog bool   `opt:"auto_clean_task_log" default:"true"`
	Hostname         string `opt:"hostname" default:""`
	CfgPath          string
	iniFile          *ini.File
	createConfigFile bool
	DriverName       string `opt:"driver_name" default:"sqlite3"`
	DSN              string `opt:"dsn" default:"data/jiacrontabd.db"`
}

func (c *Config) Resolve() error {
	c.iniFile = c.loadConfig(c.CfgPath)
	if c.createConfigFile {
		defer c.iniFile.SaveTo(c.CfgPath)
	}

	val := reflect.ValueOf(c).Elem()
	typ := val.Type()
	hostname := util.GetHostname()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		opt := field.Tag.Get("opt")
		if opt == "" {
			continue
		}
		sec := c.iniFile.Section("jiacrontabd")

		if opt == "hostname" && sec.Key(opt).String() == "" {
			val.Field(i).SetString(hostname)
		}

		if c.createConfigFile {
			key := sec.Key(opt)
			key.Comment = field.Tag.Get("comment")
			key.SetValue(fmt.Sprint(val.Field(i).Interface()))
		}

		if !sec.HasKey(opt) {
			continue
		}

		key := sec.Key(opt)
		switch field.Type.Kind() {
		case reflect.Bool:
			v, err := key.Bool()
			if err != nil {
				log.Error(err)
			}
			val.Field(i).SetBool(v)
		case reflect.String:
			val.Field(i).SetString(key.String())

		}
	}
	return nil
}

func NewConfig() *Config {
	return &Config{
		LogLevel:         "warn",
		VerboseJobLog:    true,
		ListenAddr:       "127.0.0.1:20001",
		LocalAddr:        "127.0.0.1:20002",
		AdminAddr:        "127.0.0.1:20003",
		LogPath:          "./logs",
		PprofAddr:        "127.0.0.1:20004",
		AutoCleanTaskLog: true,
		CfgPath:          "./jiacrontabd.ini",
		DriverName:       "sqlite3",
		DSN:              "data/jiacrontabd.db",
	}
}

func (c *Config) loadConfig(path string) *ini.File {
	if !file.Exist(path) {
		f, err := file.CreateFile(path)
		if err != nil {
			panic(err)
		}
		f.Close()
		c.createConfigFile = true
	}

	iniFile, err := ini.Load(path)
	if err != nil {
		panic(err)
	}
	return iniFile
}
